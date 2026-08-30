package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createColorRequest struct {
	Name     string `json:"name" binding:"required"`
	HexValue string `json:"hex_value" binding:"required,hexcolor"`
}

type listColorsRequest struct {
	PageSize  int32  `form:"page_size" binding:"omitempty,min=5,max=100"`
	PageID    int32  `form:"page_id" binding:"min=0"`
	Search    string `form:"search"`
	SortBy    string `form:"sort_by,default=created_at" binding:"oneof=id name hex_value created_at"`
	SortOrder string `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

func (server *Server) createColor(ctx *gin.Context) {
	var req createColorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateColorParams{
		Name:     req.Name,
		HexValue: req.HexValue,
	}

	color, err := server.store.CreateColor(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, color)
}

func (server *Server) listColors(ctx *gin.Context) {
	var req listColorsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if req.PageSize == 0 {
		colors, err := server.store.ListColors(ctx)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"colors": colors})
		return
	}

	colors, err := server.store.ListColorsPage(ctx, db.ListColorsPageParams{
		Search: req.Search, SortBy: req.SortBy, SortOrder: req.SortOrder,
		PageSize: req.PageSize, PageOffset: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountColors(ctx, req.Search)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"colors": colors, "total": total})
}

type updateColorRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Name     string `json:"name" binding:"required"`
	HexValue string `json:"hex_value" binding:"required,hexcolor"`
}

func (server *Server) updateColor(ctx *gin.Context) {
	var req updateColorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	color, err := server.store.UpdateColor(ctx, db.UpdateColorParams{
		ID:       req.ID,
		Name:     req.Name,
		HexValue: req.HexValue,
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
	ctx.JSON(http.StatusOK, color)
}

type deleteColorRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteColor(ctx *gin.Context) {
	var req deleteColorRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteColor(ctx, req.ID)
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
