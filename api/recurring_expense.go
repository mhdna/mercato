package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createRecurringExpenseRequest struct {
	Description  string `json:"description" binding:"required"`
	Category     string `json:"category"`
	Amount       int64  `json:"amount" binding:"required"`
	CurrencyCode string `json:"currency_code" binding:"required"`
	// IntervalUnit is "day" or "month" -- e.g. electricity every 5 months is
	// {interval_unit: "month", interval_count: 5}.
	IntervalUnit  string    `json:"interval_unit" binding:"required,oneof=day month"`
	IntervalCount int32     `json:"interval_count" binding:"required,min=1"`
	NextDueAt     time.Time `json:"next_due_at" binding:"required"`
}

func (server *Server) createRecurringExpense(ctx *gin.Context) {
	var req createRecurringExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	recurring, err := server.store.CreateRecurringExpense(ctx, db.CreateRecurringExpenseParams{
		Description:   req.Description,
		Category:      req.Category,
		Amount:        req.Amount,
		CurrencyCode:  req.CurrencyCode,
		IntervalUnit:  req.IntervalUnit,
		IntervalCount: req.IntervalCount,
		NextDueAt:     req.NextDueAt,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recurring)
}

func (server *Server) listRecurringExpenses(ctx *gin.Context) {
	recurring, err := server.store.ListRecurringExpenses(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recurring)
}

type setRecurringExpenseActiveRequest struct {
	ID     int64 `json:"id" binding:"required,min=1"`
	Active bool  `json:"active"`
}

// setRecurringExpenseActive pauses/resumes a template without touching its
// schedule -- deleting it would lose next_due_at's position in the
// interval, which a "pause electricity while we're between suppliers"
// admin action shouldn't do.
func (server *Server) setRecurringExpenseActive(ctx *gin.Context) {
	var req setRecurringExpenseActiveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	recurring, err := server.store.SetRecurringExpenseActive(ctx, db.SetRecurringExpenseActiveParams{
		ID:     req.ID,
		Active: req.Active,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recurring)
}
