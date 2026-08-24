package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type posSettingsRequest struct {
	CashboxID int64 `form:"cashbox_id" binding:"required,min=1"`
}

func (server *Server) posSettings(ctx *gin.Context) {
	var req posSettingsRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	salespersons, err := server.store.ListSalespersonsByCashbox(ctx, sql.NullInt64{
		Int64: req.CashboxID,
		Valid: true,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	cashboxAccounts, err := server.store.ListAllCashboxAccounts(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	currencies, err := server.store.ListAllCurrencies(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"salespersons":     salespersons,
		"cashbox_accounts": cashboxAccounts,
		"currencies":       currencies,
	})
}
