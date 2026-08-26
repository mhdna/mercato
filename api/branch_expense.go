package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchExpenseRequest is the exact JSON body POSTed to /branch/expenses by
// kashi-pos (see branchExpensePayload in that repo's sync_outbox.go, which
// must stay in sync with this shape). BranchCashboxAccountID/BranchShiftID
// are the branch's own local SQLite ids, meaningless to kashi's schema --
// stored as opaque traceability data, not foreign keys, same reasoning as
// branchInvoiceRequest in branch_sync.go.
//
// Category is free text, resolved to expense_categories.category_id via
// get-or-create (see getOrCreateExpenseCategoryID below) when non-blank --
// kashi-pos never fetches or caches kashi's category list, it just reports
// whatever text was typed in, same as branchLoanRequest.CategoryName.
type branchExpenseRequest struct {
	ClientRef              string `json:"client_ref" binding:"required"`
	Description            string `json:"description" binding:"required"`
	Category               string `json:"category"`
	Amount                 int64  `json:"amount" binding:"required"`
	CurrencyCode           string `json:"currency_code" binding:"required"`
	BranchCashboxAccountID int64  `json:"branch_cashbox_account_id" binding:"required"`
	// A branch may report an expense with no shift open (e.g. before the
	// day's first shift starts), so this is optional unlike
	// BranchCashboxAccountID.
	BranchShiftID *int64    `json:"branch_shift_id"`
	OccurredAt    time.Time `json:"occurred_at" binding:"required"`
}

// createBranchExpense follows the same idempotent-insert pattern as
// createBranchInvoice in branch_sync.go: look up by (branch_id, client_ref)
// first so a retried sync_outbox entry is a no-op, and fall back to the
// same lookup on a unique-violation race.
func (server *Server) createBranchExpense(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if existing, err := server.store.GetBranchExpenseByClientRef(ctx, db.GetBranchExpenseByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.ClientRef,
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"expense": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	var categoryID sql.NullInt64
	if req.Category != "" {
		id, err := server.getOrCreateExpenseCategoryID(ctx, req.Category)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		categoryID = sql.NullInt64{Int64: id, Valid: true}
	}

	var branchShiftID sql.NullInt64
	if req.BranchShiftID != nil {
		branchShiftID = sql.NullInt64{Int64: *req.BranchShiftID, Valid: true}
	}

	expense, err := server.store.CreateBranchExpense(ctx, db.CreateBranchExpenseParams{
		BranchID:               branchID,
		ClientRef:              req.ClientRef,
		Description:            req.Description,
		CategoryID:             categoryID,
		Amount:                 req.Amount,
		CurrencyCode:           req.CurrencyCode,
		BranchCashboxAccountID: req.BranchCashboxAccountID,
		BranchShiftID:          branchShiftID,
		OccurredAt:             req.OccurredAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchExpenseByClientRef(ctx, db.GetBranchExpenseByClientRefParams{
				BranchID:  branchID,
				ClientRef: req.ClientRef,
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"expense": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Only the genuinely-new-insert path reaches here (the lookup above
	// returns early on "already exists"), same convention as
	// createBranchInvoice in branch_sync.go. Amount is negated: an expense
	// is always money leaving, so the toast queue can treat it the same as
	// a return -- red, no separate "kind" field needed.
	server.adminHub.broadcastAll(adminWSMessage{
		Type:         "branch_expense_created",
		BranchID:     branchID,
		Amount:       -req.Amount,
		CurrencyCode: req.CurrencyCode,
		Label:        req.Description,
	})

	// Mint a one-time receipt-upload token for this new expense so
	// kashi-pos can show a QR code for it immediately -- see
	// expense_upload.go. Only done on the fresh-insert path: a retried
	// outbox entry that hits the "already exists" branches above must
	// never mint a second token for the same expense.
	uploadToken, err := server.createBranchExpenseUploadToken(ctx, expense.ID)
	if err != nil {
		// The expense itself is safely created; losing the upload token is
		// a degraded experience (no QR dialog), not a failed request.
		log.Printf("create upload token for branch expense %d: %v", expense.ID, err)
		server.writeJSON(ctx, http.StatusOK, envelope{"expense": expense})
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"expense": expense, "upload_token": uploadToken})
}

type listBranchExpensesRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	BranchID int64 `form:"branch_id"`
}

// listBranchExpenses is the admin-facing counterpart to createBranchExpense,
// same pagination shape as listBranchInvoices for consistency with the rest
// of the admin UI.
func (server *Server) listBranchExpenses(ctx *gin.Context) {
	var req listBranchExpensesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	expenses, err := server.store.ListBranchExpenses(ctx, db.ListBranchExpensesParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchExpenses(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_expenses": expenses, "total": total})
}

// getOrCreateExpenseCategoryID resolves a category name reported by a
// branch to its expense_categories.id, creating the category (active by
// default) on first use. Admin CRUD (expense_category.go) remains the only
// way to rename/deactivate a category -- this path only ever adds new
// ones, it never mutates an existing row. Mirrors
// getOrCreateLoanCategoryID in branch_loan.go.
func (server *Server) getOrCreateExpenseCategoryID(ctx *gin.Context, name string) (int64, error) {
	categories, err := server.store.ListExpenseCategories(ctx)
	if err != nil {
		return 0, err
	}
	for _, category := range categories {
		if category.Name == name {
			return category.ID, nil
		}
	}

	category, err := server.store.CreateExpenseCategory(ctx, db.CreateExpenseCategoryParams{
		Name:     name,
		IsActive: true,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.ListExpenseCategories(ctx); lookupErr == nil {
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
