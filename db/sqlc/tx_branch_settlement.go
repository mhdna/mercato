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

type ApplyBranchSettlementTxResult struct {
	Settlement     BranchInvoiceSettlement
	Payments       []BranchInvoiceSettlementPayment
	InvoiceUpdated bool
}

// ApplyBranchSettlementTx records a branch "change settlement" edit and, if
// the re-settled sale is already on the cloud, replays the new per-account
// split onto that sale's branch_invoice_payments so the admin view stays
// correct. A settlement whose sale hasn't synced yet is still recorded (the
// settlement + its payment lines); only the branch-invoice replay is
// skipped. The whole thing is one transaction so a partial replay can never
// leave the sale's payments half-rewritten.
func (store *SQLStore) ApplyBranchSettlementTx(ctx context.Context, arg ApplyBranchSettlementTxParams) (ApplyBranchSettlementTxResult, error) {
	var result ApplyBranchSettlementTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		settlement, err := q.CreateBranchInvoiceSettlement(ctx, CreateBranchInvoiceSettlementParams{
			BranchID:      arg.BranchID,
			ClientRef:     arg.ClientRef,
			SaleClientRef: arg.SaleClientRef,
			GrandTotal:    arg.GrandTotal,
			OccurredAt:    arg.OccurredAt,
		})
		if err != nil {
			return err
		}
		result.Settlement = settlement

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

		invoice, err := q.GetBranchInvoiceByClientRef(ctx, GetBranchInvoiceByClientRefParams{
			BranchID:  arg.BranchID,
			ClientRef: arg.SaleClientRef,
		})
		if errors.Is(err, sql.ErrNoRows) {
			// The sale isn't on the cloud yet -- nothing to replay onto.
			return nil
		}
		if err != nil {
			return err
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
