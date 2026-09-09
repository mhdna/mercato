package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type purchaseItemRequest struct {
	VariantID *int64 `json:"variant_id"`
	AssetID   *int64 `json:"asset_id"`
	Quantity  int64  `json:"quantity" binding:"required"`
	UnitPrice int64  `json:"unit_price"`
}

type createPurchaseRequest struct {
	SupplierID   int64                 `json:"supplier_id" binding:"required"`
	InventoryID  int64                 `json:"inventory_id" binding:"required"`
	Code         string                `json:"code"`
	CurrencyCode string                `json:"currency_code" binding:"required"`
	PurchasedAt  string                `json:"purchased_at"`
	Note         string                `json:"note"`
	Items        []purchaseItemRequest `json:"items" binding:"omitempty,dive"`
}

// createPurchase records a draft purchase invoice and its lines in one
// call. Nothing enters stock until the purchase is received.
func (server *Server) createPurchase(ctx *gin.Context) {
	var req createPurchaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var purchasedAt sql.NullTime
	if req.PurchasedAt != "" {
		t, err := time.Parse(time.RFC3339, req.PurchasedAt)
		if err != nil {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		purchasedAt = sql.NullTime{Time: t, Valid: true}
	}

	items := make([]db.PurchaseItemParams, 0, len(req.Items))
	for _, it := range req.Items {
		p := db.PurchaseItemParams{Quantity: it.Quantity, UnitPrice: it.UnitPrice}
		if it.VariantID != nil {
			p.VariantID = sql.NullInt64{Int64: *it.VariantID, Valid: true}
		}
		if it.AssetID != nil {
			p.AssetID = sql.NullInt64{Int64: *it.AssetID, Valid: true}
		}
		items = append(items, p)
	}

	result, err := server.store.CreatePurchaseTx(ctx, db.CreatePurchaseTxParams{
		SupplierID:   req.SupplierID,
		InventoryID:  req.InventoryID,
		Code:         req.Code,
		CurrencyCode: req.CurrencyCode,
		PurchasedAt:  purchasedAt,
		Note:         req.Note,
		Items:        items,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"purchase": result.Purchase, "items": result.Items})
}

type getPurchaseRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getPurchase(ctx *gin.Context) {
	var req getPurchaseRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	purchase, err := server.store.GetPurchase(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	itemRows, err := server.store.ListPurchaseItems(ctx, sql.NullInt64{Int64: req.ID, Valid: true})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"purchase": purchase, "items": itemRows})
}

type listPurchasesRequest struct {
	PageSize int32  `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
	Search   string `form:"search"`
}

func (server *Server) listPurchases(ctx *gin.Context) {
	var req listPurchasesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	purchases, err := server.store.ListPurchases(ctx, db.ListPurchasesParams{
		Search:     req.Search,
		PageSize:   req.PageSize,
		PageOffset: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountPurchasesFiltered(ctx, req.Search)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"purchases": purchases, "total": total})
}

type addPurchaseItemRequest struct {
	PurchaseID   int64  `json:"purchase_id" binding:"required"`
	VariantID    *int64 `json:"variant_id"`
	AssetID      *int64 `json:"asset_id"`
	Quantity     int64  `json:"quantity" binding:"required"`
	UnitPrice    int64  `json:"unit_price"`
	CurrencyCode string `json:"currency_code" binding:"required"`
}

// addPurchaseItem appends a line to an existing draft purchase and
// refreshes the header totals.
func (server *Server) addPurchaseItem(ctx *gin.Context) {
	var req addPurchaseItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.AddPurchaseItemParams{
		PurchaseID:   sql.NullInt64{Int64: req.PurchaseID, Valid: true},
		Quantity:     req.Quantity,
		UnitPrice:    req.UnitPrice,
		CurrencyCode: req.CurrencyCode,
	}
	if req.VariantID != nil {
		arg.VariantID = sql.NullInt64{Int64: *req.VariantID, Valid: true}
	}
	if req.AssetID != nil {
		arg.AssetID = sql.NullInt64{Int64: *req.AssetID, Valid: true}
	}

	item, err := server.store.AddPurchaseItem(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := server.refreshPurchaseTotals(ctx, req.PurchaseID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (server *Server) refreshPurchaseTotals(ctx *gin.Context, purchaseID int64) error {
	rows, err := server.store.ListPurchaseItems(ctx, sql.NullInt64{Int64: purchaseID, Valid: true})
	if err != nil {
		return err
	}
	var subtotal int64
	for _, r := range rows {
		subtotal += r.UnitPrice * r.Quantity
	}
	return server.store.SetPurchaseTotals(ctx, db.SetPurchaseTotalsParams{
		ID:         purchaseID,
		Subtotal:   subtotal,
		GrandTotal: subtotal,
	})
}

type purchaseIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// receivePurchase moves a draft purchase to 'received': its lines land in
// the destination inventory and roll each SKU's moving-average cost.
func (server *Server) receivePurchase(ctx *gin.Context) {
	var req purchaseIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.PurchaseReceiveTx(ctx, db.PurchaseReceiveTxParams{
		PurchaseID: req.ID,
		CreatedBy:  actorID(ctx),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"purchase": result.Purchase, "movements": result.Movements})
}

// cancelPurchase moves a purchase to 'cancelled'. A received purchase is
// unwound with compensating negative movements first.
func (server *Server) cancelPurchase(ctx *gin.Context) {
	var req purchaseIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.CancelPurchaseTx(ctx, req.ID, actorID(ctx))
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"purchase": result.Purchase, "movements": result.Movements})
}
