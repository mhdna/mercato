package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type transferItemRequest struct {
	VariantID *int64 `json:"variant_id"`
	AssetID   *int64 `json:"asset_id"`
	Quantity  int64  `json:"quantity" binding:"required"`
}

type createTransferRequest struct {
	FromInventoryID int64                 `json:"from_inventory_id" binding:"required"`
	ToInventoryID   int64                 `json:"to_inventory_id" binding:"required"`
	Type            string                `json:"type"`
	Code            string                `json:"code"`
	Note            string                `json:"note"`
	Items           []transferItemRequest `json:"items" binding:"omitempty,dive"`
}

// createTransfer records a draft transfer and its lines. Stock only moves
// on dispatch (out of source) and receive (into destination).
func (server *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	transferType := db.TransferTypeProducts
	if req.Type != "" {
		transferType = db.TransferType(req.Type)
	}

	items := make([]db.TransferItemParams, 0, len(req.Items))
	for _, it := range req.Items {
		p := db.TransferItemParams{Quantity: it.Quantity}
		if it.VariantID != nil {
			p.VariantID = sql.NullInt64{Int64: *it.VariantID, Valid: true}
		}
		if it.AssetID != nil {
			p.AssetID = sql.NullInt64{Int64: *it.AssetID, Valid: true}
		}
		items = append(items, p)
	}

	result, err := server.store.CreateTransferTx(ctx, db.CreateTransferTxParams{
		FromInventoryID: req.FromInventoryID,
		ToInventoryID:   req.ToInventoryID,
		Type:            transferType,
		Code:            req.Code,
		Note:            req.Note,
		Items:           items,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"transfer": result.Transfer, "items": result.Items})
}

type transferIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getTransfer(ctx *gin.Context) {
	var req transferIDRequest
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

	items, err := server.store.ListTransferItems(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"transfer": transfer, "items": items})
}

type listTransfersRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listTransfers(ctx *gin.Context) {
	var req listTransfersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	transfers, err := server.store.ListTransfers(ctx, db.ListTransfersParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	})
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
	Note            string `json:"note"`
}

func (server *Server) updateTransfer(ctx *gin.Context) {
	var req updateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.UpdateTransfer(ctx, db.UpdateTransferParams{
		ID:              req.ID,
		FromInventoryID: req.FromInventoryID,
		ToInventoryID:   req.ToInventoryID,
		Type:            db.TransferType(req.Type),
		Note:            req.Note,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "transfer updated"})
}

type createTransferItemRequest struct {
	TransferID int64  `json:"transfer_id" binding:"required"`
	VariantID  *int64 `json:"variant_id"`
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
		Quantity:   req.Quantity,
	}
	if req.VariantID != nil {
		arg.VariantID = sql.NullInt64{Int64: *req.VariantID, Valid: true}
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

// dispatchTransfer removes the transfer's lines from the source inventory
// (stock goes "in transit").
func (server *Server) dispatchTransfer(ctx *gin.Context) {
	var req transferIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.TransferDispatchTx(ctx, db.TransferStageTxParams{TransferID: req.ID, CreatedBy: actorID(ctx)})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"transfer": result.Transfer, "movements": result.Movements})
}

// receiveTransfer lands the transfer's lines in the destination inventory.
func (server *Server) receiveTransfer(ctx *gin.Context) {
	var req transferIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.TransferReceiveTx(ctx, db.TransferStageTxParams{TransferID: req.ID, CreatedBy: actorID(ctx)})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"transfer": result.Transfer, "movements": result.Movements})
}

// cancelTransfer moves a transfer to 'cancelled', unwinding whichever legs
// already ran.
func (server *Server) cancelTransfer(ctx *gin.Context) {
	var req transferIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.CancelTransferTx(ctx, req.ID, actorID(ctx))
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"transfer": result.Transfer, "movements": result.Movements})
}
