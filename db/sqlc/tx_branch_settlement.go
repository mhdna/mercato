package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ApplyBranchSettlementTxParams struct {
	BranchID      int64
	ClientRef     string
	SaleClientRef string
	GrandTotal    int64
	OccurredAt    time.Time
	Payments      []BranchInvoicePaymentParams
}

// Settlement lifecycle -- see db/migrations/000075.
const (
	SettlementStatusApplied            = "applied"
	SettlementStatusPending            = "pending"
	SettlementStatusIgnoredShiftClosed = "ignored_shift_closed"
)

type ApplyBranchSettlementTxResult struct {
	Settlement     BranchInvoiceSettlement
	Payments       []BranchInvoiceSettlementPayment
	InvoiceUpdated bool
	// Status is the resolved lifecycle of this settlement (one of the
	// SettlementStatus* constants). Frozen is the shorthand the caller
	// reacts to: the sale's shift was already closed, so nothing was
	// replayed and the change is deliberately not surfaced.
	Status string
	Frozen bool
}

// ApplyBranchSettlementTx records a branch "change settlement" edit and, if
// the re-settled sale is already on the cloud AND still belongs to the
// branch's open shift, replays the new per-account split onto that sale's
// branch_invoice_payments so the admin view stays correct.
//
// Three cases, all recorded (settlement row + its payment lines, for the
// audit trail) so a retried outbox entry always gets its 200:
//
//   - sale on cloud, its shift still open  -> replayed  (status "applied")
//   - sale not on cloud yet                -> not replayed (status "pending")
//   - sale on cloud, its shift already closed -> not replayed, and the
//     caller is told Frozen so it doesn't emit a "settlement changed"
//     notification (status "ignored_shift_closed")
//
// The whole thing is one transaction so a partial replay can never leave
// the sale's payments half-rewritten.
func (store *SQLStore) ApplyBranchSettlementTx(ctx context.Context, arg ApplyBranchSettlementTxParams) (ApplyBranchSettlementTxResult, error) {
	var result ApplyBranchSettlementTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// Resolve the target sale (and whether its shift is closed) BEFORE
		// inserting, so the settlement row is stamped with its final status.
		invoice, err := q.GetBranchInvoiceByClientRef(ctx, GetBranchInvoiceByClientRefParams{
			BranchID:  arg.BranchID,
			ClientRef: arg.SaleClientRef,
		})
		saleOnCloud := true
		if errors.Is(err, sql.ErrNoRows) {
			saleOnCloud = false
		} else if err != nil {
			return err
		}

		status := SettlementStatusPending
		if saleOnCloud {
			closed, err := q.BranchShiftIsClosed(ctx, BranchShiftIsClosedParams{
				BranchID:      arg.BranchID,
				BranchShiftID: invoice.BranchShiftID,
			})
			if err != nil {
				return err
			}
			if closed {
				status = SettlementStatusIgnoredShiftClosed
			} else {
				status = SettlementStatusApplied
			}
		}

		settlement, err := q.CreateBranchInvoiceSettlement(ctx, CreateBranchInvoiceSettlementParams{
			BranchID:      arg.BranchID,
			ClientRef:     arg.ClientRef,
			SaleClientRef: arg.SaleClientRef,
			GrandTotal:    arg.GrandTotal,
			OccurredAt:    arg.OccurredAt,
			Status:        status,
		})
		if err != nil {
			return err
		}
		result.Settlement = settlement
		result.Status = status
		result.Frozen = status == SettlementStatusIgnoredShiftClosed

		result.Payments = make([]BranchInvoiceSettlementPayment, 0, len(arg.Payments))
		for _, payment := range arg.Payments {
			created, err := q.CreateBranchInvoiceSettlementPayment(ctx, CreateBranchInvoiceSettlementPaymentParams{
				BranchInvoiceSettlementID: settlement.ID,
				AccountName:               payment.AccountName,
				Amount:                    payment.Amount,
			})
			if err != nil {
				return err
			}
			result.Payments = append(result.Payments, created)
		}

		if status != SettlementStatusApplied {
			return nil
		}

		if err := q.DeleteBranchInvoicePaymentsForInvoice(ctx, invoice.ID); err != nil {
			return err
		}
		for _, payment := range arg.Payments {
			if _, err := q.CreateBranchInvoicePayment(ctx, CreateBranchInvoicePaymentParams{
				BranchInvoiceID: invoice.ID,
				AccountName:     payment.AccountName,
				Amount:          payment.Amount,
			}); err != nil {
				return err
			}
		}
		result.InvoiceUpdated = true
		return nil
	})

	return result, err
}
