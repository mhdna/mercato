package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createTransferRequest struct {
	FromInventoryID int64  `json:"from_inventory_id" binding:"required"`
	ToInventoryID   int64  `json:"to_inventory_id" binding:"required"`
	Type            string `json:"type" binding:"required"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateTransferParams{
		FromInventoryID: req.FromInventoryID,
		ToInventoryID:   req.ToInventoryID,
		Type:            db.TransferType(req.Type),
	}

	transfer, err := server.store.CreateTransfer(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}

type getTransferRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	transfer, err := server.store.GetTransfer(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

type listTransfersRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listTransfers(ctx *gin.Context) {
	var req listTransfersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListTransfersParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	transfers, err := server.store.ListTransfers(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountTransfers(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"transfers": transfers, "total": total})
}

type updateTransferRequest struct {
	ID              int64  `json:"id" binding:"required,min=1"`
	FromInventoryID int64  `json:"from_inventory_id" binding:"required"`
	ToInventoryID   int64  `json:"to_inventory_id" binding:"required"`
	Type            string `json:"type" binding:"required"`
}

func (server *Server) updateTransfer(ctx *gin.Context) {
	var req updateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.UpdateTransferParams{
		ID:              req.ID,
		FromInventoryID: req.FromInventoryID,
		ToInventoryID:   req.ToInventoryID,
		Type:            db.TransferType(req.Type),
	}
	err := server.store.UpdateTransfer(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "transfer updated"})
}

type createTransferItemRequest struct {
	TransferID int64  `json:"transfer_id" binding:"required"`
	ProductID  *int64 `json:"product_id"`
	AssetID    *int64 `json:"asset_id"`
	Quantity   int64  `json:"quantity" binding:"required"`
}

func (server *Server) createTransferItem(ctx *gin.Context) {
	var req createTransferItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateTransferItemParams{
		TransferID: req.TransferID,
		ProductID:  sql.NullInt64{Int64: 0, Valid: false},
		AssetID:    sql.NullInt64{Int64: 0, Valid: false},
		Quantity:   req.Quantity,
	}
	if req.ProductID != nil {
		arg.ProductID = sql.NullInt64{Int64: *req.ProductID, Valid: true}
	}
	if req.AssetID != nil {
		arg.AssetID = sql.NullInt64{Int64: *req.AssetID, Valid: true}
	}

	item, err := server.store.CreateTransferItem(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, item)
}

type listTransferItemsRequest struct {
	TransferID int64 `uri:"transfer_id" binding:"required,min=1"`
}

func (server *Server) listTransferItems(ctx *gin.Context) {
	var req listTransferItemsRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	items, err := server.store.ListTransferItems(ctx, req.TransferID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, items)
}
