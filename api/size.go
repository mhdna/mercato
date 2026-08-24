package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createSizeRequest struct {
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required"`
	Order string `json:"order" binding:"required"`
}

func (server *Server) createSize(ctx *gin.Context) {
	var req createSizeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateSizeParams{
		Name:  req.Name,
		Type:  req.Type,
		Order: req.Order,
	}

	size, err := server.store.CreateSize(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, size)
}

func (server *Server) listSizes(ctx *gin.Context) {
	sizes, err := server.store.ListSizes(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sizes": sizes})
}
