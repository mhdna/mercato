package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createEntryRequest struct {
	CashboxID     int64  `json:"cashbox_id" binding:"required"`
	InventoryID   int64  `json:"inventory_id" binding:"required"`
	ReferenceType string `json:"reference_type" binding:"required"`
	ReferenceID   int64  `json:"reference_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
}

func (server *Server) createEntry(ctx *gin.Context) {
	var req createEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateEntryItemParams{
		CashboxID:     req.CashboxID,
		InventoryID:   req.InventoryID,
		ReferenceType: db.EntryReferenceType(req.ReferenceType),
		ReferenceID:   req.ReferenceID,
		Amount:        req.Amount,
	}

	entry, err := server.store.CreateEntryItem(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, entry)
}

type getEntryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getEntry(ctx *gin.Context) {
	var req getEntryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	entry, err := server.store.GetEntry(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type listEntriesRequest struct {
	InventoryID int64 `form:"inventory_id" binding:"required,min=1"`
	PageSize    int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID      int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var req listEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListEntriesParams{
		InventoryID: req.InventoryID,
		Limit:       req.PageSize,
		Offset:      req.PageID,
	}
	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entries)
}
