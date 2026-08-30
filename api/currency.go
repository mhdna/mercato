package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// defaultUnitsPerUSDMicros/defaultCashRoundingUnit match the values
// kashi-pos itself defaults to for a currency at par with USD (see its
// 000052_currency_rates.up.sql) — used here when a caller doesn't specify
// them, so existing callers written before these fields existed keep
// getting sane behavior.
const (
	defaultUnitsPerUSDMicros = 1_000_000
	defaultCashRoundingUnit  = 1
)

type createCurrencyRequest struct {
	Name                   string `json:"name" binding:"required"`
	Code                   string `json:"code" binding:"required"`
	Symbol                 string `json:"symbol" binding:"required"`
	ValueInDefaultCurrency int64  `json:"value_in_default_currency" binding:"required,min=1"`
	UnitsPerUsdMicros      int64  `json:"units_per_usd_micros"`
	CashRoundingUnit       int64  `json:"cash_rounding_unit"`
}

func (server *Server) createCurrency(ctx *gin.Context) {
	var req createCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.UnitsPerUsdMicros == 0 {
		req.UnitsPerUsdMicros = defaultUnitsPerUSDMicros
	}
	if req.CashRoundingUnit == 0 {
		req.CashRoundingUnit = defaultCashRoundingUnit
	}

	arg := db.CreateCurrencyParams{
		Name:                   req.Name,
		Code:                   req.Code,
		Symbol:                 req.Symbol,
		ValueInDefaultCurrency: req.ValueInDefaultCurrency,
		UnitsPerUsdMicros:      req.UnitsPerUsdMicros,
		CashRoundingUnit:       req.CashRoundingUnit,
	}

	currency, err := server.store.CreateCurrency(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "currency_updated"})
	ctx.JSON(http.StatusOK, currency)
}

type getCurrencyRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) getCurrency(ctx *gin.Context) {
	var req getCurrencyRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	currency, err := server.store.GetCurrency(ctx, req.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, currency)
}

type listCurrencies struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	// all=true returns every currency unpaginated -- used by the currency
	// picker (CurrencySelect.vue), which needs the whole list, not a page.
	All bool `form:"all"`
}

func (server *Server) listCurrencies(ctx *gin.Context) {
	var req listCurrencies
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	// all=true -> page_size 0 -> LIMIT NULLIF(0,0) -> every row, unpaginated.
	pageSize := req.PageSize
	if req.All {
		pageSize = 0
	}

	respondList(server, ctx, "currencies",
		func() ([]db.Currency, error) {
			return server.store.ListCurrencies(ctx, db.ListCurrenciesParams{
				PageSize: pageSize, PageOffset: req.PageID,
			})
		},
		func() (int64, error) { return server.store.CountCurrencies(ctx) },
	)
}

type updateCurrencyRequest struct {
	Code                   string `json:"code" binding:"required"`
	Name                   string `json:"name" binding:"required"`
	Symbol                 string `json:"symbol" binding:"required"`
	ValueInDefaultCurrency int64  `json:"value_in_default_currency" binding:"required,min=1"`
	IsActive               *bool  `json:"is_active"`
	UnitsPerUsdMicros      int64  `json:"units_per_usd_micros"`
	CashRoundingUnit       int64  `json:"cash_rounding_unit"`
}

func (server *Server) updateCurrency(ctx *gin.Context) {
	var req updateCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.UnitsPerUsdMicros == 0 {
		req.UnitsPerUsdMicros = defaultUnitsPerUSDMicros
	}
	if req.CashRoundingUnit == 0 {
		req.CashRoundingUnit = defaultCashRoundingUnit
	}

	currency, err := server.store.UpdateCurrency(ctx, db.UpdateCurrencyParams{
		Code:                   req.Code,
		Name:                   req.Name,
		Symbol:                 req.Symbol,
		ValueInDefaultCurrency: req.ValueInDefaultCurrency,
		IsActive:               isActive,
		UnitsPerUsdMicros:      req.UnitsPerUsdMicros,
		CashRoundingUnit:       req.CashRoundingUnit,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "currency_updated"})
	ctx.JSON(http.StatusOK, currency)
}

type deleteCurrencyRequest struct {
	Code string `uri:"code" binding:"required"`
}

func (server *Server) deleteCurrency(ctx *gin.Context) {
	var req deleteCurrencyRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteCurrency(ctx, req.Code)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
