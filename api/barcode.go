package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

// listBarcodes lists product variants for the barcodes browse/print page.
// Each variant's barcode is a real per-SKU barcode (see product_variants,
// which replaced the old one-barcode-per-product table).
type listBarcodeRequest struct {
	PageSize int32  `form:"page_size,default=15" binding:"min=5,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
	Search   string `form:"search"`
}

func (server *Server) listBarcodes(ctx *gin.Context) {
	var req listBarcodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	variants, err := server.store.ListVariantsForBarcodes(ctx, db.ListVariantsForBarcodesParams{
		Search:     req.Search,
		PageLimit:  req.PageSize,
		PageOffset: req.PageID * req.PageSize,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountVariantsForBarcodes(ctx, req.Search)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"variants": variants, "total": total})
}

// assignBarcodes gives a fresh, unique EAN-13 barcode to any selected
// variant whose current barcode isn't a valid EAN-13 (e.g. legacy/imported
// rows). Variants that already have a valid barcode are left untouched.
type assignBarcodesRequest struct {
	VariantIDs []int64 `json:"variant_ids" binding:"required,min=1"`
}

func (server *Server) assignBarcodes(ctx *gin.Context) {
	var req assignBarcodesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	updated := make([]db.ProductVariant, 0)
	skipped := 0

	for _, id := range req.VariantIDs {
		variant, err := server.store.GetProductVariant(ctx, id)
		if err != nil {
			server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("variant %d: %w", id, err))
			return
		}
		if isEAN13(variant.Barcode) {
			skipped++
			continue
		}

		newVariant, err := server.assignUniqueBarcode(ctx, id)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		updated = append(updated, newVariant)
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"updated": updated, "skipped": skipped})
}

func (server *Server) assignUniqueBarcode(ctx *gin.Context, variantID int64) (db.ProductVariant, error) {
	for attempt := 0; attempt < 20; attempt++ {
		itemNumber, err := server.store.GetNextBarcodeItemValue(ctx)
		if err != nil {
			return db.ProductVariant{}, err
		}
		code := generateEAN13Str(itemNumber)

		// Guard against a code that somehow already exists on another row.
		if _, err := server.store.GetProductVariantByBarcode(ctx, code); err == nil {
			continue
		}

		v, err := server.store.SetVariantBarcode(ctx, db.SetVariantBarcodeParams{ID: variantID, Barcode: code})
		if err == nil {
			return v, nil
		}
		var pqErr *pq.Error
		if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
			return db.ProductVariant{}, err
		}
	}
	return db.ProductVariant{}, errors.New("failed to generate a unique barcode after 20 attempts")
}

// generateEAN13Str mirrors db.generateEAN13 but returns the string form.
func generateEAN13Str(itemNumber int64) string {
	base := fmt.Sprintf("800%09d", itemNumber)
	sum := 0
	for i, ch := range base {
		d := int(ch - '0')
		if i%2 == 0 {
			sum += d * 3
		} else {
			sum += d
		}
	}
	check := (10 - sum%10) % 10
	return base + strconv.Itoa(check)
}

// printBarcodes renders a single PDF of barcode labels for the selected
// variants, with an optional separator/header page before each item.
type printBarcodesRequest struct {
	Items []struct {
		VariantID int64 `json:"variant_id" binding:"required"`
		Qty       int   `json:"qty"`
	} `json:"items" binding:"required,min=1"`
	Separator struct {
		Enabled bool     `json:"enabled"`
		Fields  []string `json:"fields"`
	} `json:"separator"`
	Label struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"label"`
}

func (server *Server) printBarcodes(ctx *gin.Context) {
	var req printBarcodesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	width, height := req.Label.Width, req.Label.Height
	if width <= 0 || height <= 0 {
		settings, err := server.store.GetAppSettings(ctx)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		width, height = float64(settings.BarcodeLabelWidth), float64(settings.BarcodeLabelHeight)
	}

	fields := map[string]bool{}
	for _, f := range req.Separator.Fields {
		fields[f] = true
	}

	items := make([]labelItem, 0, len(req.Items))
	for _, in := range req.Items {
		v, err := server.store.GetVariantForLabel(ctx, in.VariantID)
		if err != nil {
			server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("variant %d: %w", in.VariantID, err))
			return
		}
		if !isEAN13(v.Barcode) {
			server.writeError(ctx, http.StatusBadRequest,
				fmt.Errorf("variant %d has no valid barcode yet -- run Generate barcodes first", in.VariantID))
			return
		}
		price := ""
		if v.Price.Valid {
			price = fmt.Sprintf("%d.%02d", v.Price.Int64/100, v.Price.Int64%100)
		}
		qty := in.Qty
		if qty < 1 {
			qty = 1
		}
		items = append(items, labelItem{
			Name:    v.ProductName,
			Code:    v.ProductCode,
			Barcode: v.Barcode,
			Price:   price,
			Brand:   v.Brand,
			Color:   v.ColorName,
			Size:    v.SizeName,
			Qty:     qty,
		})
	}

	pdf, err := buildLabelsPDF(items, labelOptions{
		Width:     width,
		Height:    height,
		Separator: req.Separator.Enabled,
		Fields:    fields,
	})
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Header("Content-Disposition", `inline; filename="barcodes.pdf"`)
	ctx.Data(http.StatusOK, "application/pdf", pdf)
}
