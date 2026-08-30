package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createExpenseRequest struct {
	Description  string `json:"description" binding:"required"`
	CategoryID   int64  `json:"category_id"`
	Amount       int64  `json:"amount" binding:"required"`
	CurrencyCode string `json:"currency_code" binding:"required"`
}

func (server *Server) createExpense(ctx *gin.Context) {
	var req createExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var categoryID sql.NullInt64
	if req.CategoryID > 0 {
		categoryID = sql.NullInt64{Int64: req.CategoryID, Valid: true}
	}

	arg := db.CreateExpenseParams{
		Description:  req.Description,
		CategoryID:   categoryID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		// Manually created through this endpoint, not fired by a recurring
		// template -- see recurring_expense.go for the other path into
		// CreateExpense.
		RecurringExpenseID: sql.NullInt64{},
	}

	expense, err := server.store.CreateExpense(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, expense)
}

type getExpenseRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getExpense(ctx *gin.Context) {
	var req getExpenseRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	expense, err := server.store.GetExpense(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, expense)
}

type listExpensesRequest struct {
	PageSize int32 `form:"page_size,default=25" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listExpenses(ctx *gin.Context) {
	var req listExpensesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListExpensesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	expenses, err := server.store.ListExpenses(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, expenses)
}

type updateExpenseRequest struct {
	ID           int64  `json:"id" binding:"required,min=1"`
	Description  string `json:"description" binding:"required"`
	CategoryID   int64  `json:"category_id"`
	Amount       int64  `json:"amount" binding:"required"`
	CurrencyCode string `json:"currency_code" binding:"required"`
}

func (server *Server) updateExpense(ctx *gin.Context) {
	var req updateExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var categoryID sql.NullInt64
	if req.CategoryID > 0 {
		categoryID = sql.NullInt64{Int64: req.CategoryID, Valid: true}
	}

	expense, err := server.store.UpdateExpense(ctx, db.UpdateExpenseParams{
		ID:           req.ID,
		Description:  req.Description,
		CategoryID:   categoryID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, expense)
}

type deleteExpenseRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteExpense(ctx *gin.Context) {
	var req deleteExpenseRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteExpense(ctx, req.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
