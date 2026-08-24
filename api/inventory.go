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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	inventory, err := server.store.GetInventory(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListInventoriesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	inventories, err := server.store.ListInventories(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	total, err := server.store.CountInventories(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.UpdateInventory(ctx, db.UpdateInventoryParams{
		ID:   req.ID,
		Name: req.Name,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteInventory(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
