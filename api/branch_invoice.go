package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type listBranchInvoicesRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	BranchID int64 `form:"branch_id"`
}

// listBranchInvoices is the admin-facing counterpart to branch_sync.go's
// createBranchInvoice: that endpoint is how a branch's data gets in, this
// is how a human admin looks at it, same pagination shape as
// listSalesInvoices (sales_invoice.go) for consistency with the rest of the
// admin UI.
func (server *Server) listBranchInvoices(ctx *gin.Context) {
	var req listBranchInvoicesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	invoices, err := server.store.ListBranchInvoices(ctx, db.ListBranchInvoicesParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchInvoices(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_invoices": invoices, "total": total})
}

type branchInvoiceIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) listBranchInvoiceItems(ctx *gin.Context) {
	var req branchInvoiceIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.GetBranchInvoice(ctx, req.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	items, err := server.store.ListBranchInvoiceItems(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"items": items})
}

type dailyIncomeRequest struct {
	Year     int32 `form:"year"`
	BranchID int64 `form:"branch_id"`
}

type dailyIncomeEntry struct {
	Day   time.Time `json:"day"`
	Total int64     `json:"total"`
}

// dailyIncome backs the dashboard's "Branches Daily Income" heatmap --
// grand_total per day for a year, optionally scoped to one branch.
// Empty days simply don't appear in the result; the frontend fills zeros.
func (server *Server) dailyIncome(ctx *gin.Context) {
	var req dailyIncomeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.Year == 0 {
		req.Year = int32(time.Now().Year())
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	rows, err := server.store.ListDailyIncome(ctx, db.ListDailyIncomeParams{
		Year:     req.Year,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	entries := make([]dailyIncomeEntry, len(rows))
	for i, row := range rows {
		entries[i] = dailyIncomeEntry{Day: row.Day, Total: row.Total}
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"year": req.Year, "days": entries})
}
