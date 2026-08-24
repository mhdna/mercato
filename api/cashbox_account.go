package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createCashboxAccountRequest struct {
	Name         string `json:"name" binding:"required"`
	CurrencyCode string `json:"currency_code"`
	SortOrder    int32  `json:"sort_order"`
	Color        string `json:"color"`
}

func (server *Server) createCashboxAccount(ctx *gin.Context) {
	var req createCashboxAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.CurrencyCode == "" {
		req.CurrencyCode = "USD"
	}

	cashboxAccount, err := server.store.CreateCashboxAccount(ctx, db.CreateCashboxAccountParams{
		Name:         req.Name,
		CurrencyCode: req.CurrencyCode,
		SortOrder:    req.SortOrder,
		Color:        req.Color,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "cashbox_account_updated"})
	ctx.JSON(http.StatusOK, cashboxAccount)
}

type listCashboxAccountsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listCashboxAccounts(ctx *gin.Context) {
	var req listCashboxAccountsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListCashboxAccountsParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	cashboxAccounts, err := server.store.ListCashboxAccounts(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cashboxAccounts)
}

type updateCashboxAccountRequest struct {
	ID           int64  `json:"id" binding:"required,min=1"`
	Name         string `json:"name" binding:"required"`
	CurrencyCode string `json:"currency_code"`
	SortOrder    int32  `json:"sort_order"`
	Color        string `json:"color"`
}

func (server *Server) updateCashboxAccount(ctx *gin.Context) {
	var req updateCashboxAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.CurrencyCode == "" {
		req.CurrencyCode = "USD"
	}

	arg := db.UpdateCashboxAccountParams{
		ID:           req.ID,
		Name:         req.Name,
		CurrencyCode: req.CurrencyCode,
		SortOrder:    req.SortOrder,
		Color:        req.Color,
	}
	cashboxAccount, err := server.store.UpdateCashboxAccount(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "cashbox_account_updated"})
	ctx.JSON(http.StatusOK, cashboxAccount)
}

type addCashboxAccountBalance struct {
	AccountID int64 `json:"account_id" binding:"required,min=1"`
	ShiftID   int64 `json:"shift_id" binding:"required,min=1"`
	Amount    int64 `json:"amount" binding:"required"`
}

func (server *Server) addCashboxAccountBalance(ctx *gin.Context) {
	var req addCashboxAccountBalance
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.AddCashboxAccountBalanceParams{
		AccountID: req.AccountID,
		Balance:   req.Amount,
		ShiftID:   req.ShiftID,
	}
	cashboxBalance, err := server.store.AddCashboxAccountBalance(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, cashboxBalance)
}
