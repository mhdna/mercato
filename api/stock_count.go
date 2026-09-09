package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type stockCountLineRequest struct {
	VariantID       int64 `json:"variant_id" binding:"required"`
	CountedQuantity int64 `json:"counted_quantity"`
}

type saveStockCountRequest struct {
	InventoryID int64                   `json:"inventory_id" binding:"required"`
	Note        string                  `json:"note"`
	Items       []stockCountLineRequest `json:"items"`
}

// createStockCount starts a new draft count sheet.
func (server *Server) createStockCount(ctx *gin.Context) {
	var req saveStockCountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	count, err := server.store.SaveStockCountTx(ctx, db.SaveStockCountTxParams{
		InventoryID: req.InventoryID,
		Note:        req.Note,
		CreatedBy:   actorID(ctx),
		Items:       toStockCountLines(req.Items),
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, count)
}

type updateStockCountRequest struct {
	ID          int64                   `json:"id" binding:"required,min=1"`
	InventoryID int64                   `json:"inventory_id" binding:"required"`
	Note        string                  `json:"note"`
	Items       []stockCountLineRequest `json:"items"`
}

// updateStockCount rewrites a draft count's lines (re-snapshotting on-hand).
func (server *Server) updateStockCount(ctx *gin.Context) {
	var req updateStockCountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	count, err := server.store.SaveStockCountTx(ctx, db.SaveStockCountTxParams{
		ID:          req.ID,
		InventoryID: req.InventoryID,
		Note:        req.Note,
		CreatedBy:   actorID(ctx),
		Items:       toStockCountLines(req.Items),
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, count)
}

func toStockCountLines(items []stockCountLineRequest) []db.StockCountLineParams {
	lines := make([]db.StockCountLineParams, 0, len(items))
	for _, it := range items {
		lines = append(lines, db.StockCountLineParams{
			VariantID:       it.VariantID,
			CountedQuantity: it.CountedQuantity,
		})
	}
	return lines
}

type stockCountIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getStockCount(ctx *gin.Context) {
	var req stockCountIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	count, err := server.store.GetStockCount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	items, err := server.store.ListStockCountItems(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"stock_count": count, "items": items})
}

type listStockCountsRequest struct {
	InventoryID int64  `form:"inventory_id"`
	Status      string `form:"status"`
	PageSize    int32  `form:"page_size,default=14" binding:"min=1,max=100"`
	PageID      int32  `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listStockCounts(ctx *gin.Context) {
	var req listStockCountsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var inventoryID sql.NullInt64
	if req.InventoryID > 0 {
		inventoryID = sql.NullInt64{Int64: req.InventoryID, Valid: true}
	}
	var status db.NullStockCountStatus
	if req.Status != "" {
		status = db.NullStockCountStatus{StockCountStatus: db.StockCountStatus(req.Status), Valid: true}
	}

	respondList(server, ctx, "stock_counts",
		func() ([]db.ListStockCountsRow, error) {
			return server.store.ListStockCounts(ctx, db.ListStockCountsParams{
				InventoryID: inventoryID,
				Status:      status,
				PageLimit:   req.PageSize,
				PageOffset:  req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountStockCounts(ctx, db.CountStockCountsParams{
				InventoryID: inventoryID,
				Status:      status,
			})
		},
	)
}

// postStockCount posts a draft count: differences against live on-hand
// become stock movements.
func (server *Server) postStockCount(ctx *gin.Context) {
	var req stockCountIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.PostStockCountTx(ctx, req.ID, actorID(ctx))
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"stock_count": result.StockCount, "movements": result.Movements})
}

func (server *Server) cancelStockCount(ctx *gin.Context) {
	var req stockCountIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	count, err := server.store.CancelStockCountTx(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, count)
}

// listInventoryCountSheet returns every SKU stocked in this inventory with
// its live on-hand, to seed a new count. Reuses ListInventoryStock.
func (server *Server) listInventoryCountSheet(ctx *gin.Context) {
	var uri getInventoryRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req listInventoryStockRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	respondList(server, ctx, "stock",
		func() ([]db.ListInventoryStockRow, error) {
			return server.store.ListInventoryStock(ctx, db.ListInventoryStockParams{
				InventoryID: uri.ID,
				Search:      req.Search,
				PageLimit:   req.PageSize,
				PageOffset:  req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountInventoryStock(ctx, db.CountInventoryStockParams{
				InventoryID: uri.ID,
				Search:      req.Search,
			})
		},
	)
}
