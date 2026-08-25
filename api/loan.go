package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createLoanRequest struct {
	Description  string `json:"description" binding:"required"`
	CategoryID   int64  `json:"category_id" binding:"required,min=1"`
	Amount       int64  `json:"amount" binding:"required"`
	CurrencyCode string `json:"currency_code" binding:"required"`
}

// createLoan is the admin-facing counterpart to createBranchLoan in
// branch_loan.go -- it writes a 'central_loan' row directly, with no
// branch/client_ref fields, same split as CreateExpense vs
// CreateBranchExpense.
func (server *Server) createLoan(ctx *gin.Context) {
	var req createLoanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	loan, err := server.store.CreateCentralLoan(ctx, db.CreateCentralLoanParams{
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, loan)
}

type getLoanRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getLoan(ctx *gin.Context) {
	var req getLoanRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	loan, err := server.store.GetLoan(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, loan)
}

type listLoansRequest struct {
	PageSize int32  `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32  `form:"page_id,default=0" binding:"min=0"`
	BranchID int64  `form:"branch_id"`
	Origin   string `form:"origin" binding:"omitempty,oneof=central_loan branch_loan"`
}

// listLoans is the single admin listing endpoint for both central- and
// branch-origin loans, filterable by either -- unlike expenses, which
// splits ListExpenses/ListBranchExpenses across two tables, loans live in
// one table so one endpoint covers both origins.
func (server *Server) listLoans(ctx *gin.Context) {
	var req listLoansRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}
	var origin sql.NullString
	if req.Origin != "" {
		origin = sql.NullString{String: req.Origin, Valid: true}
	}

	loans, err := server.store.ListLoans(ctx, db.ListLoansParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
		Origin:   origin,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountLoans(ctx, db.CountLoansParams{
		BranchID: branchID,
		Origin:   origin,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"loans": loans, "total": total})
}
