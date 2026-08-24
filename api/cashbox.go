package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createCashboxRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (server *Server) createCashbox(ctx *gin.Context) {
	var req createCashboxRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateCashboxParams{
		Code: req.Code,
		Name: req.Name,
	}

	cashbox, err := server.store.CreateCashbox(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, cashbox)
}

type getCashboxRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getCashbox(ctx *gin.Context) {
	var req getCashboxRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	cashbox, err := server.store.GetCashbox(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, cashbox)
}

type listCashboxesRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listCashboxes(ctx *gin.Context) {
	var req listCashboxesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCashboxesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	cashboxes, err := server.store.ListCashboxes(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, cashboxes)
}

type updateCashboxRequest struct {
	ID   int64  `json:"id" binding:"required,min=1"`
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (server *Server) updateCashbox(ctx *gin.Context) {
	var req updateCashboxRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateCashboxParams{
		ID:   req.ID,
		Name: req.Name,
		Code: req.Code,
	}
	cashbox, err := server.store.UpdateCashbox(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, cashbox)
}
