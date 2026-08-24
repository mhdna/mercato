package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createCurrencyRequest struct {
	Name                   string `json:"name" binding:"required"`
	Code                   string `json:"code" binding:"required"`
	Symbol                 string `json:"symbol" binding:"required"`
	ValueInDefaultCurrency int64  `json:"value_in_default_currency" binding:"required,min=1"`
}

func (server *Server) createCurrency(ctx *gin.Context) {
	var req createCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateCurrencyParams{
		Name:                   req.Name,
		Code:                   req.Code,
		Symbol:                 req.Symbol,
		ValueInDefaultCurrency: req.ValueInDefaultCurrency,
	}

	currency, err := server.store.CreateCurrency(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, currency)
}

type getCurrencyRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) getCurrency(ctx *gin.Context) {
	var req getCurrencyRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	currency, err := server.store.GetCurrency(ctx, req.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, currency)
}

type listCurrencies struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listCurrencies(ctx *gin.Context) {
	var req listCurrencies
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCurrenciesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	currencies, err := server.store.ListCurrencies(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	total, err := server.store.CountCurrencies(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"currencies": currencies, "total": total})
}

type updateCurrencyRequest struct {
	Code                   string `json:"code" binding:"required"`
	Name                   string `json:"name" binding:"required"`
	Symbol                 string `json:"symbol" binding:"required"`
	ValueInDefaultCurrency int64  `json:"value_in_default_currency" binding:"required,min=1"`
}

func (server *Server) updateCurrency(ctx *gin.Context) {
	var req updateCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	currency, err := server.store.UpdateCurrency(ctx, db.UpdateCurrencyParams{
		Code:                   req.Code,
		Name:                   req.Name,
		Symbol:                 req.Symbol,
		ValueInDefaultCurrency: req.ValueInDefaultCurrency,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, currency)
}

type deleteCurrencyRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) deleteCurrency(ctx *gin.Context) {
	var req deleteCurrencyRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteCurrency(ctx, req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
