package db

import "context"

type BranchInvoiceItemParams struct {
	BranchProductID int64 `json:"branch_product_id"`
	UnitPrice       int64 `json:"unit_price"`
	LineTotal       int64 `json:"line_total"`
	Discount        int16 `json:"discount"`
	Quantity        int64 `json:"quantity"`
}

type CreateBranchInvoiceTxParams struct {
	CreateBranchInvoiceParams
	Items []BranchInvoiceItemParams `json:"items"`
}

type CreateBranchInvoiceTxResult struct {
	Invoice BranchInvoice       `json:"invoice"`
	Items   []BranchInvoiceItem `json:"items"`
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

		result.Invoice = invoice
		result.Items = items
		return nil
	})

	return result, err
}
