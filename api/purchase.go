package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createPurchaseRequest struct {
	SupplierID  int64  `json:"supplier_id" binding:"required"`
	PurchasedAt string `json:"purchased_at" binding:"required"`
}

func (server *Server) createPurchase(ctx *gin.Context) {
	var req createPurchaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	purchasedAt, err := time.Parse(time.RFC3339, req.PurchasedAt)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreatePurchaseParams{
		SupplierID:  req.SupplierID,
		PurchasedAt: purchasedAt,
	}

	purchase, err := server.store.CreatePurchase(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, purchase)
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

	ctx.JSON(http.StatusOK, purchase)
}

type listPurchasesRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listPurchases(ctx *gin.Context) {
	var req listPurchasesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListPurchasesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	purchases, err := server.store.ListPurchases(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountPurchases(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"purchases": purchases, "total": total})
}

type addPurchaseItemRequest struct {
	PurchaseID   *int64 `json:"purchase_id"`
	ProductID    *int64 `json:"product_id"`
	AssetID      *int64 `json:"asset_id"`
	Quantity     int64  `json:"quantity" binding:"required"`
	UnitPrice    int64  `json:"unit_price" binding:"required"`
	CurrencyCode string `json:"currency_code" binding:"required"`
}

func (server *Server) addPurchaseItem(ctx *gin.Context) {
	var req addPurchaseItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.AddPurchaseItemParams{
		PurchaseID:   sql.NullInt64{Int64: 0, Valid: false},
		ProductID:    sql.NullInt64{Int64: 0, Valid: false},
		AssetID:      sql.NullInt64{Int64: 0, Valid: false},
		Quantity:     req.Quantity,
		UnitPrice:    req.UnitPrice,
		CurrencyCode: req.CurrencyCode,
	}
	if req.PurchaseID != nil {
		arg.PurchaseID = sql.NullInt64{Int64: *req.PurchaseID, Valid: true}
	}
	if req.ProductID != nil {
		arg.ProductID = sql.NullInt64{Int64: *req.ProductID, Valid: true}
	}
	if req.AssetID != nil {
		arg.AssetID = sql.NullInt64{Int64: *req.AssetID, Valid: true}
	}

	item, err := server.store.AddPurchaseItem(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, item)
}
