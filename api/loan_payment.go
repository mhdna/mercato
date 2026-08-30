package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createLoanPaymentRequest struct {
	LoanID       int64  `json:"loan_id" binding:"required,min=1"`
	Amount       int64  `json:"amount" binding:"required"`
	CurrencyCode string `json:"currency_code" binding:"required"`
	Note         string `json:"note"`
}

// createLoanPayment records a repayment (سحوبات) against a loan. There is
// no branch-facing counterpart -- loan payments are entered centrally
// only, against a loan regardless of whether that loan itself originated
// centrally or from a branch.
func (server *Server) createLoanPayment(ctx *gin.Context) {
	var req createLoanPaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.GetLoan(ctx, req.LoanID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	payment, err := server.store.CreateLoanPayment(ctx, db.CreateLoanPaymentParams{
		LoanID:       req.LoanID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		Note:         req.Note,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, payment)
}

type listLoanPaymentsRequest struct {
	PageSize   int32  `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID     int32  `form:"page_id,default=0" binding:"min=0"`
	LoanID     int64  `form:"loan_id"`
	CategoryID int64  `form:"category_id"`
	Search     string `form:"search"`
}

func (server *Server) listLoanPayments(ctx *gin.Context) {
	var req listLoanPaymentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var loanID sql.NullInt64
	if req.LoanID > 0 {
		loanID = sql.NullInt64{Int64: req.LoanID, Valid: true}
	}
	var categoryID sql.NullInt64
	if req.CategoryID > 0 {
		categoryID = sql.NullInt64{Int64: req.CategoryID, Valid: true}
	}
	var search sql.NullString
	if req.Search != "" {
		search = sql.NullString{String: req.Search, Valid: true}
	}

	payments, err := server.store.ListLoanPayments(ctx, db.ListLoanPaymentsParams{
		Limit:      req.PageSize,
		Offset:     req.PageID,
		LoanID:     loanID,
		CategoryID: categoryID,
		Search:     search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountLoanPayments(ctx, db.CountLoanPaymentsParams{
		LoanID:     loanID,
		CategoryID: categoryID,
		Search:     search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"loan_payments": payments, "total": total})
}

type deleteLoanPaymentRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteLoanPayment(ctx *gin.Context) {
	var req deleteLoanPaymentRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteLoanPayment(ctx, req.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
