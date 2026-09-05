package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createInventoryRequest struct {
	Name      string           `json:"name" binding:"required"`
	Type      db.InventoryType `json:"type" binding:"required"`
	Code      string           `json:"code" binding:"required"`
	Longitude *float64         `json:"longitude"`
	Latitude  *float64         `json:"latitude"`
}

func (server *Server) createInventory(ctx *gin.Context) {
	var req createInventoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var longitude, latitude sql.NullFloat64
	if req.Longitude != nil {
		longitude = sql.NullFloat64{Float64: *req.Longitude, Valid: true}
	}
	if req.Latitude != nil {
		latitude = sql.NullFloat64{Float64: *req.Latitude, Valid: true}
	}

	arg := db.CreateInventoryParams{
		Name:      req.Name,
		Type:      req.Type,
		Code:      req.Code,
		Longitude: longitude,
		Latitude:  latitude,
	}
	inventory, err := server.store.CreateInventory(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, inventory)
}

type getInventoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getInventory(ctx *gin.Context) {
	var req getInventoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	inventory, err := server.store.GetInventory(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, inventory)
}

type listInventoryRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listInventories(ctx *gin.Context) {
	var req listInventoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListInventoriesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	inventories, err := server.store.ListInventories(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountInventories(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"inventories": inventories, "total": total})
}

type updateInventoryRequest struct {
	ID   int64  `json:"id" binding:"required,min=1"`
	Name string `json:"name" binding:"required"`
}

func (server *Server) updateInventory(ctx *gin.Context) {
	var req updateInventoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.UpdateInventory(ctx, db.UpdateInventoryParams{
		ID:   req.ID,
		Name: req.Name,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}

type deleteInventoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteInventory(ctx *gin.Context) {
	var req deleteInventoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteInventory(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ---------------------------------------------------------------------------
// Per-variant stock, adjustments, and the movement ledger
// ---------------------------------------------------------------------------

type listInventoryStockRequest struct {
	Search   string `form:"search"`
	PageSize int32  `form:"page_size,default=50" binding:"min=1,max=500"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
}

// listInventoryStock returns the on-hand quantity and moving-average cost
// of every SKU that has ever been stocked in this inventory.
func (server *Server) listInventoryStock(ctx *gin.Context) {
	var uri getInventoryRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	// Bind the query separately: the paging fields rely on their `default`
	// tags, which only apply during form binding. Validating them as part of
	// the URI bind would reject every request because they're still zero.
	var req listInventoryStockRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	rows, err := server.store.ListInventoryStock(ctx, db.ListInventoryStockParams{
		InventoryID: uri.ID,
		Search:      req.Search,
		PageLimit:   req.PageSize,
		PageOffset:  req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountInventoryStock(ctx, db.CountInventoryStockParams{
		InventoryID: uri.ID,
		Search:      req.Search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"stock": rows, "total": total})
}

type createStockAdjustmentRequest struct {
	VariantID int64  `json:"variant_id" binding:"required"`
	Mode      string `json:"mode" binding:"required,oneof=delta count"`
	Quantity  int64  `json:"quantity"`
	Note      string `json:"note"`
}

// createStockAdjustment applies a manual correction to one SKU's on-hand
// in this inventory: "delta" adds a signed quantity, "count" sets on-hand
// to an absolute counted figure. Either way one balancing movement is
// written -- no supporting document.
func (server *Server) createStockAdjustment(ctx *gin.Context) {
	var uri getInventoryRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req createStockAdjustmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.StockAdjustmentTx(ctx, db.StockAdjustmentTxParams{
		InventoryID: uri.ID,
		VariantID:   req.VariantID,
		Mode:        db.StockAdjustmentMode(req.Mode),
		Quantity:    req.Quantity,
		Note:        req.Note,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"movement": result.Movement, "on_hand": result.OnHand})
}

type listStockMovementsRequest struct {
	InventoryID int64  `form:"inventory_id"`
	VariantID   int64  `form:"variant_id"`
	Reason      string `form:"reason"`
	PageSize    int32  `form:"page_size,default=50" binding:"min=1,max=500"`
	PageID      int32  `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listStockMovements(ctx *gin.Context) {
	var req listStockMovementsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var inventoryID, variantID sql.NullInt64
	if req.InventoryID > 0 {
		inventoryID = sql.NullInt64{Int64: req.InventoryID, Valid: true}
	}
	if req.VariantID > 0 {
		variantID = sql.NullInt64{Int64: req.VariantID, Valid: true}
	}
	var reason db.NullStockMovementReason
	if req.Reason != "" {
		reason = db.NullStockMovementReason{StockMovementReason: db.StockMovementReason(req.Reason), Valid: true}
	}

	rows, err := server.store.ListStockMovements(ctx, db.ListStockMovementsParams{
		InventoryID: inventoryID,
		VariantID:   variantID,
		Reason:      reason,
		PageLimit:   req.PageSize,
		PageOffset:  req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountStockMovements(ctx, db.CountStockMovementsParams{
		InventoryID: inventoryID,
		VariantID:   variantID,
		Reason:      reason,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"movements": rows, "total": total})
}
