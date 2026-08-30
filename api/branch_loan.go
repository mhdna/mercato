package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchLoanRequest is the exact JSON body POSTed to /branch/loans by
// kashi-pos (see branchLoanPayload in that repo's sync_outbox.go, which
// must stay in sync with this shape). BranchCashboxAccountID/BranchShiftID
// are the branch's own local SQLite ids, meaningless to kashi's schema --
// stored as opaque traceability data, not foreign keys, same reasoning as
// branchExpenseRequest in branch_expense.go.
//
// CategoryName resolves to loan_categories.category_id via get-or-create
// (see getOrCreateLoanCategoryID below) rather than the branch sending a
// category_id directly -- kashi-pos never fetches or caches kashi's
// category list, it just reports the lender-source name it was given.
type branchLoanRequest struct {
	ClientRef              string `json:"client_ref" binding:"required"`
	Description            string `json:"description" binding:"required"`
	CategoryName           string `json:"category_name" binding:"required"`
	Amount                 int64  `json:"amount" binding:"required"`
	CurrencyCode           string `json:"currency_code" binding:"required"`
	BranchCashboxAccountID int64  `json:"branch_cashbox_account_id" binding:"required"`
	// A branch may report a loan with no shift open, same reasoning as
	// BranchShiftID on branchExpenseRequest.
	BranchShiftID *int64    `json:"branch_shift_id"`
	OccurredAt    time.Time `json:"occurred_at" binding:"required"`
}

// createBranchLoan follows the same idempotent-insert pattern as
// createBranchExpense: look up by (branch_id, client_ref) first so a
// retried sync_outbox entry is a no-op, and fall back to the same lookup
// on a unique-violation race.
func (server *Server) createBranchLoan(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchLoanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if existing, err := server.store.GetBranchLoanByClientRef(ctx, db.GetBranchLoanByClientRefParams{
		BranchID:  sql.NullInt64{Int64: branchID, Valid: true},
		ClientRef: sql.NullString{String: req.ClientRef, Valid: true},
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"loan": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	categoryID, err := server.getOrCreateLoanCategoryID(ctx, req.CategoryName)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	var branchShiftID sql.NullInt64
	if req.BranchShiftID != nil {
		branchShiftID = sql.NullInt64{Int64: *req.BranchShiftID, Valid: true}
	}

	loan, err := server.store.CreateBranchLoan(ctx, db.CreateBranchLoanParams{
		BranchID:               sql.NullInt64{Int64: branchID, Valid: true},
		ClientRef:              sql.NullString{String: req.ClientRef, Valid: true},
		Description:            req.Description,
		CategoryID:             categoryID,
		Amount:                 req.Amount,
		CurrencyCode:           req.CurrencyCode,
		BranchCashboxAccountID: sql.NullInt64{Int64: req.BranchCashboxAccountID, Valid: true},
		BranchShiftID:          branchShiftID,
		OccurredAt:             req.OccurredAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchLoanByClientRef(ctx, db.GetBranchLoanByClientRefParams{
				BranchID:  sql.NullInt64{Int64: branchID, Valid: true},
				ClientRef: sql.NullString{String: req.ClientRef, Valid: true},
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"loan": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Only the genuinely-new-insert path reaches here (the lookup above
	// returns early on "already exists"), same convention as
	// createBranchExpense. Amount is positive: a loan is money coming in,
	// same sign convention as an invoice/sale on the toast queue.
	server.adminHub.broadcastAll(adminWSMessage{
		Type:         "branch_loan_created",
		BranchID:     branchID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		Label:        req.Description,
	})

	server.writeJSON(ctx, http.StatusOK, envelope{"loan": loan})
}

// getOrCreateLoanCategoryID resolves a lender-source name reported by a
// branch to its loan_categories.id, creating the category (active,
// scope='branch' so it syncs back down to kashi-pos) on first use. Admin
// CRUD (loan_category.go) remains the only way to rename/deactivate a
// category -- this path only ever adds new ones, it never mutates an
// existing row.
func (server *Server) getOrCreateLoanCategoryID(ctx *gin.Context, name string) (int64, error) {
	categories, err := server.store.ListLoanCategories(ctx, sql.NullString{})
	if err != nil {
		return 0, err
	}
	for _, category := range categories {
		if category.Name == name {
			return category.ID, nil
		}
	}

	category, err := server.store.CreateLoanCategory(ctx, db.CreateLoanCategoryParams{
		Name:     name,
		IsActive: true,
		Icon:     defaultLoanCategoryIcon,
		Color:    defaultLoanCategoryColor,
		Scope:    "branch",
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.ListLoanCategories(ctx, sql.NullString{}); lookupErr == nil {
				for _, category := range existing {
					if category.Name == name {
						return category.ID, nil
					}
				}
			}
		}
		return 0, err
	}
	return category.ID, nil
}
