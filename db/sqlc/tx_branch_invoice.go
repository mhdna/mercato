package db

import (
	"context"
	"database/sql"
	"errors"
)

type BranchInvoiceItemParams struct {
	BranchProductID int64 `json:"branch_product_id"`
	// Barcode is the SKU's globally-unique barcode -- the only identifier
	// that resolves a branch line to a central product_variants row (see
	// applySyncedProductVariant in kashi-pos). Empty from older kashi-pos
	// builds; such lines are recorded but skip the central stock decrement.
	Barcode   string `json:"barcode"`
	UnitPrice int64  `json:"unit_price"`
	LineTotal int64  `json:"line_total"`
	Discount  int16  `json:"discount"`
	Quantity  int64  `json:"quantity"`
}

type BranchInvoicePaymentParams struct {
	AccountName string `json:"account_name"`
	Amount      int64  `json:"amount"`
}

type CreateBranchInvoiceTxParams struct {
	CreateBranchInvoiceParams
	Items    []BranchInvoiceItemParams    `json:"items"`
	Payments []BranchInvoicePaymentParams `json:"payments"`
}

type CreateBranchInvoiceTxResult struct {
	Invoice  BranchInvoice          `json:"invoice"`
	Items    []BranchInvoiceItem    `json:"items"`
	Payments []BranchInvoicePayment `json:"payments"`
	// StockSkippedItems counts line items whose barcode did not resolve to
	// a central variant (or whose branch has no linked inventory); those
	// lines are stored but did not move central stock.
	StockSkippedItems int `json:"stock_skipped_items"`
}

// CreateBranchInvoiceTx inserts a branch invoice header and its line items
// together, so a crash mid-write can never leave an idempotency-key row
// with no items behind (which would make a legitimate retry think the
// invoice was already fully recorded).
func (store *SQLStore) CreateBranchInvoiceTx(ctx context.Context, arg CreateBranchInvoiceTxParams) (CreateBranchInvoiceTxResult, error) {
	var result CreateBranchInvoiceTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		invoice, err := q.CreateBranchInvoice(ctx, arg.CreateBranchInvoiceParams)
		if err != nil {
			return err
		}

		items := make([]BranchInvoiceItem, 0, len(arg.Items))
		for _, item := range arg.Items {
			created, err := q.CreateBranchInvoiceItem(ctx, CreateBranchInvoiceItemParams{
				BranchInvoiceID: invoice.ID,
				BranchProductID: item.BranchProductID,
				UnitPrice:       item.UnitPrice,
				LineTotal:       item.LineTotal,
				Discount:        item.Discount,
				Quantity:        item.Quantity,
			})
			if err != nil {
				return err
			}
			items = append(items, created)
		}

		payments := make([]BranchInvoicePayment, 0, len(arg.Payments))
		for _, payment := range arg.Payments {
			created, err := q.CreateBranchInvoicePayment(ctx, CreateBranchInvoicePaymentParams{
				BranchInvoiceID: invoice.ID,
				AccountName:     payment.AccountName,
				Amount:          payment.Amount,
			})
			if err != nil {
				return err
			}
			payments = append(payments, created)
		}

		// Decrement (or, for a return, restore) central stock for this
		// branch's own store inventory, resolving each line by barcode. A
		// line that can't be resolved -- unknown barcode, missing barcode,
		// or a branch with no linked inventory -- is left recorded but
		// untouched in stock, never a hard failure (an offline branch
		// catching up must always be able to land its invoices).
		branch, err := q.GetBranch(ctx, arg.BranchID)
		if err != nil {
			return err
		}
		reason := StockMovementReasonSale
		if arg.Kind == "return" {
			reason = StockMovementReasonReturn
		}
		for _, item := range arg.Items {
			if !branch.InventoryID.Valid || item.Barcode == "" {
				result.StockSkippedItems++
				continue
			}
			variant, err := q.GetProductVariantByBarcode(ctx, item.Barcode)
			if errors.Is(err, sql.ErrNoRows) {
				result.StockSkippedItems++
				continue
			}
			if err != nil {
				return err
			}
			if _, err := q.applyStockMovement(ctx, applyStockMovementParams{
				InventoryID:   branch.InventoryID.Int64,
				VariantID:     variant.ID,
				Quantity:      -item.Quantity,
				Reason:        reason,
				ReferenceType: "branch_invoice",
				ReferenceID:   invoice.ID,
			}); err != nil {
				return err
			}
		}

		result.Invoice = invoice
		result.Items = items
		result.Payments = payments
		return nil
	})

	return result, err
}
