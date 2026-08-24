package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createColorRequest struct {
	Name     string `json:"name" binding:"required"`
	HexValue string `json:"hex_value" binding:"required,hexcolor"`
}

func (server *Server) createColor(ctx *gin.Context) {
	var req createColorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateColorParams{
		Name:     req.Name,
		HexValue: req.HexValue,
	}

	color, err := server.store.CreateColor(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, color)
}

func (server *Server) listColors(ctx *gin.Context) {
	colors, err := server.store.ListColors(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"colors": colors})
}
