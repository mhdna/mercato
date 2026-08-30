package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createSizeRequest struct {
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required"`
	Order string `json:"order" binding:"required"`
}

type listSizesRequest struct {
	PageSize  int32  `form:"page_size" binding:"omitempty,min=5,max=100"`
	PageID    int32  `form:"page_id" binding:"min=0"`
	Search    string `form:"search"`
	SortBy    string `form:"sort_by,default=created_at" binding:"oneof=id name type order created_at"`
	SortOrder string `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

func (server *Server) createSize(ctx *gin.Context) {
	var req createSizeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateSizeParams{
		Name:  req.Name,
		Type:  req.Type,
		Order: req.Order,
	}

	size, err := server.store.CreateSize(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, size)
}

func (server *Server) listSizes(ctx *gin.Context) {
	var req listSizesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if req.PageSize == 0 {
		sizes, err := server.store.ListSizes(ctx)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"sizes": sizes})
		return
	}

	sizes, err := server.store.ListSizesPage(ctx, db.ListSizesPageParams{
		Search: req.Search, SortBy: req.SortBy, SortOrder: req.SortOrder,
		PageSize: req.PageSize, PageOffset: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountSizes(ctx, req.Search)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sizes": sizes, "total": total})
}

type updateSizeRequest struct {
	ID    int64  `json:"id" binding:"required,min=1"`
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required"`
	Order string `json:"order" binding:"required"`
}

func (server *Server) updateSize(ctx *gin.Context) {
	var req updateSizeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	size, err := server.store.UpdateSize(ctx, db.UpdateSizeParams{
		ID:    req.ID,
		Name:  req.Name,
		Type:  req.Type,
		Order: req.Order,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, size)
}

type deleteSizeRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteSize(ctx *gin.Context) {
	var req deleteSizeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteSize(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "foreign_key_violation" {
			server.writeError(ctx, http.StatusConflict, errInUse)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
