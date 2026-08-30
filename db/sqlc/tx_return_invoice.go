package db

import (
	"context"
	"database/sql"
	"errors"
)

type ReturnInvoiceItem struct {
	VariantID int64 `json:"variant_id"`
	UnitPrice int64 `json:"unit_price"`
	LineTotal int64 `json:"line_total"`
	Discount  int16 `json:"discount"`
	Quantity  int64 `json:"quantity"`
}

type ReturnInvoiceTxParams struct {
	CashboxID        int64         `json:"cashbox_id"`
	ShiftID          int64         `json:"shift_id"`
	Year             int32         `json:"year"`
	ClientID         int64         `json:"client_id"`
	InventoryID      int64         `json:"inventory_id"`
	Discount         int16         `json:"discount"`
	SubTotal         int64         `json:"sub_total"`
	DiscountedTotal  int64         `json:"discounted_total"`
	GrandTotal       int64         `json:"grand_total"`
	SalesInvoiceID   int64         `json:"sales_invoice_id"`
	CashboxAccountID int64         `json:"cashbox_account_id"`
	PriceListID      sql.NullInt64 `json:"price_list_id"`
	SalespersonID    sql.NullInt64 `json:"salesperson_id"`

	Items []ReturnInvoiceItem `json:"items"`
}

type ReturnInvoiceTxResult struct {
	Invoice               Invoice               `json:"invoice"`
	SalesInvoiceID        int64                 `json:"sales_invoice_id"`
	Entry                 Entry                 `json:"entry"`
	ShiftsAccountsBalance ShiftsAccountsBalance `json:"shift_account_balance"`
}

func validateReturnInvoiceAmounts(items []ReturnInvoiceItem, grandTotal int64) error {
	var calculatedNetAmount int64
	for _, item := range items {
		itemTotal := item.UnitPrice * item.Quantity
		itemDiscountedTotal := itemTotal
		if item.Discount > 0 {
			itemDiscountedTotal = itemTotal - (itemTotal * int64(item.Discount) / 100)
		}
		calculatedNetAmount += itemDiscountedTotal
	}
	if calculatedNetAmount != grandTotal {
		return errors.New("Invalid Amount")
	}
	return nil
}

func (store *SQLStore) ReturnInvoiceTx(ctx context.Context, arg ReturnInvoiceTxParams) (ReturnInvoiceTxResult, error) {
	var result ReturnInvoiceTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		salesInvoice, err := q.GetInvoice(ctx, arg.SalesInvoiceID)
		if err != nil {
			return err
		}
		invoiceTypeID := salesInvoice.InvoiceTypeID

		invoiceIndex, err := q.generateInvoiceIndex(ctx, arg.CashboxID, arg.Year, IndexTypeReturn, invoiceTypeID)
		if err != nil {
			return err
		}

		invoiceCode, err := q.generateInvoiceNumber(ctx, EntryReferenceTypeReturnInvoice, invoiceIndex, arg.CashboxID, invoiceTypeID, arg.Year)
		if err != nil {
			return err
		}

		err = validateReturnInvoiceAmounts(arg.Items, arg.GrandTotal)
		if err != nil {
			return err
		}

		client, err := q.GetClient(ctx, arg.ClientID)
		if err != nil {
			return err
		}
		// Wholesale clients never earned points on the original sale (see
		// SalesInvoiceTx), so nothing to claw back here either.
		var loyaltyPointsDelta int64
		if client.ClientType != "wholesale" {
			loyaltyPointsDelta = -arg.GrandTotal
		}

		invoice, err := q.CreateInvoice(ctx, CreateInvoiceParams{
			CashboxID:          arg.CashboxID,
			ShiftID:            arg.ShiftID,
			InvoiceCode:        invoiceCode,
			InvoiceIndex:       invoiceIndex,
			Year:               arg.Year,
			ClientID:           arg.ClientID,
			InventoryID:        arg.InventoryID,
			Discount:           arg.Discount,
			Subtotal:           arg.SubTotal,
			DiscountedTotal:    arg.DiscountedTotal,
			GrandTotal:         arg.GrandTotal,
			InvoiceTypeID:      invoiceTypeID,
			SalespersonID:      arg.SalespersonID,
			LoyaltyPointsDelta: loyaltyPointsDelta,
		})
		if err != nil {
			return err
		}

		_, err = q.CreateReturnInvoice(ctx, CreateReturnInvoiceParams{
			InvoiceID:      invoice.ID,
			SalesInvoiceID: arg.SalesInvoiceID,
		})
		if err != nil {
			return err
		}

		for _, item := range arg.Items {
			unitPrice := item.UnitPrice
			discount := item.Discount

			variant, err := q.GetProductVariant(ctx, item.VariantID)
			if err != nil {
				return err
			}

			if arg.PriceListID.Valid {
				listPrice, priceErr := q.GetProductPriceFromList(ctx, GetProductPriceFromListParams{
					PriceListID: arg.PriceListID.Int64,
					ProductID:   variant.ProductID,
				})
				if priceErr == nil {
					unitPrice = listPrice.Price
				}
			}

			lineTotal := unitPrice * item.Quantity
			if discount > 0 {
				lineTotal = lineTotal - (lineTotal * int64(discount) / 100)
			}

			_, err = q.AddInvoiceProduct(ctx, AddInvoiceProductParams{
				InvoiceID: invoice.ID,
				ProductID: variant.ProductID,
				VariantID: sql.NullInt64{Int64: item.VariantID, Valid: true},
				UnitPrice: unitPrice,
				LineTotal: lineTotal,
				Discount:  discount,
				Quantity:  item.Quantity,
				UnitCost:  variant.AvgCost,
			})
			if err != nil {
				return err
			}
		}

		entry, err := q.CreateEntryItem(ctx, CreateEntryItemParams{
			CashboxID:     arg.CashboxID,
			InventoryID:   arg.InventoryID,
			ReferenceType: EntryReferenceTypeReturnInvoice,
			ReferenceID:   invoice.ID,
			Amount:        -arg.GrandTotal,
		})
		if err != nil {
			return err
		}

		addAccountBalanceArg := AddCashboxAccountBalanceParams{
			AccountID: arg.CashboxAccountID,
			ShiftID:   arg.ShiftID,
			Balance:   -arg.GrandTotal,
		}
		shiftAccountsBalance, err := q.AddCashboxAccountBalance(ctx, addAccountBalanceArg)
		if err != nil {
			return err
		}

		_, err = q.CreateInvoicePayment(ctx, CreateInvoicePaymentParams{
			InvoiceID:        invoice.ID,
			CashboxAccountID: arg.CashboxAccountID,
			Amount:           -arg.GrandTotal,
		})
		if err != nil {
			return err
		}

		for _, i := range arg.Items {
			_, err = q.applyStockMovement(ctx, applyStockMovementParams{
				InventoryID:   arg.InventoryID,
				VariantID:     i.VariantID,
				Quantity:      i.Quantity,
				Reason:        StockMovementReasonReturn,
				ReferenceType: "return_invoice",
				ReferenceID:   invoice.ID,
			})
			if err != nil {
				return err
			}
		}

		// Wholesale clients never earned points on the original sale (see
		// SalesInvoiceTx), so nothing to claw back here either (client was
		// already fetched above to compute loyaltyPointsDelta).
		if client.ClientType != "wholesale" {
			addPointsArg := AddClientLoyaltyPointsParams{
				ID:                 invoice.ClientID,
				TotalLoyaltyPoints: -arg.GrandTotal,
				ValidLoyaltyPoints: -arg.GrandTotal,
			}
			err = q.AddClientLoyaltyPoints(ctx, addPointsArg)
			if err != nil {
				return err
			}
		}

		result.Invoice = invoice
		result.SalesInvoiceID = arg.SalesInvoiceID
		result.Entry = entry
		result.ShiftsAccountsBalance = shiftAccountsBalance

		return nil
	})

	return result, err
}
