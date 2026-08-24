package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// listBarcodes lists product variants (each variant's barcode is a real
// per-SKU barcode now — see product_variants, which replaced the old
// one-barcode-per-product table). Kept at this URL/name since it's the
// existing UI-facing "browse barcodes" endpoint, just backed by the new
// table.
type ListBarcodeRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listBarcodes(ctx *gin.Context) {
	var req ListBarcodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	variants, err := server.store.ListProductVariants(ctx, db.ListProductVariantsParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	total, err := server.store.CountProductVariants(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"variants": variants, "total": total})
}
