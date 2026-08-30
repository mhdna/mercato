package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchSettlementPaymentRequest is one settlement line -- net amount now
// settled through one payment account. account_name is display text, not an
// id, same reasoning as branchInvoicePaymentRequest in branch_sync.go.
type branchSettlementPaymentRequest struct {
	AccountName string `json:"account_name" binding:"required"`
	Amount      int64  `json:"amount"`
}

// branchSettlementRequest is the exact JSON body POSTed to
// /branch/invoice_settlements by kashi-pos (see branchSettlementPayload in
// that repo's sync_outbox.go). It reports the new per-account payment split
// for a sale that was already synced; createBranchInvoiceSettlement replays
// it onto that sale's branch_invoice_payments so the admin view stays
// correct.
//
// SettlementRef is this edit's own idempotency key (each "change settlement"
// is a distinct event). SaleClientRef is the branch_invoices.client_ref of
// the sale being re-settled.
type branchSettlementRequest struct {
	SettlementRef string                           `json:"settlement_ref" binding:"required"`
	SaleClientRef string                           `json:"sale_client_ref" binding:"required"`
	GrandTotal    int64                            `json:"grand_total"`
	Payments      []branchSettlementPaymentRequest `json:"payments"`
	OccurredAt    time.Time                        `json:"occurred_at" binding:"required"`
}

// createBranchInvoiceSettlement follows the same idempotent-insert pattern
// as createBranchExpense: look up by (branch_id, client_ref) first so a
// retried sync_outbox entry is a no-op, and fall back to the same lookup on
// a unique-violation race.
func (server *Server) createBranchInvoiceSettlement(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchSettlementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if existing, err := server.store.GetBranchInvoiceSettlementByClientRef(ctx, db.GetBranchInvoiceSettlementByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.SettlementRef,
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"settlement": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	payments := make([]db.BranchInvoicePaymentParams, 0, len(req.Payments))
	for _, payment := range req.Payments {
		payments = append(payments, db.BranchInvoicePaymentParams{
			AccountName: payment.AccountName,
			Amount:      payment.Amount,
		})
	}

	result, err := server.store.ApplyBranchSettlementTx(ctx, db.ApplyBranchSettlementTxParams{
		BranchID:      branchID,
		ClientRef:     req.SettlementRef,
		SaleClientRef: req.SaleClientRef,
		GrandTotal:    req.GrandTotal,
		OccurredAt:    req.OccurredAt,
		Payments:      payments,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchInvoiceSettlementByClientRef(ctx, db.GetBranchInvoiceSettlementByClientRefParams{
				BranchID:  branchID,
				ClientRef: req.SettlementRef,
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"settlement": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Only the genuinely-new-insert path reaches here. The toast names the
	// sale by its branch invoice code when that sale is on the cloud,
	// falling back to the raw client_ref otherwise; Amount is the new grand
	// total, display-only.
	label := req.SaleClientRef
	if invoice, err := server.store.GetBranchInvoiceByClientRef(ctx, db.GetBranchInvoiceByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.SaleClientRef,
	}); err == nil {
		label = invoice.BranchInvoiceCode
	}
	currencyCode := ""
	if currency, err := server.store.GetDefaultCurrency(ctx); err == nil {
		currencyCode = currency.Code
	}
	server.adminHub.broadcastAll(adminWSMessage{
		Type:         "branch_settlement_changed",
		BranchID:     branchID,
		Amount:       req.GrandTotal,
		CurrencyCode: currencyCode,
		Label:        label,
	})

	server.writeJSON(ctx, http.StatusOK, envelope{
		"settlement":      result.Settlement,
		"payments":        result.Payments,
		"invoice_updated": result.InvoiceUpdated,
	})
}

type listBranchInvoiceSettlementsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	BranchID int64 `form:"branch_id"`
}

// listBranchInvoiceSettlements is the admin-facing counterpart, same
// pagination shape as listBranchExpenses.
func (server *Server) listBranchInvoiceSettlements(ctx *gin.Context) {
	var req listBranchInvoiceSettlementsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	settlements, err := server.store.ListBranchInvoiceSettlements(ctx, db.ListBranchInvoiceSettlementsParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchInvoiceSettlements(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_invoice_settlements": settlements, "total": total})
}
