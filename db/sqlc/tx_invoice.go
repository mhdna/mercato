package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type SalesInvoiceItem struct {
	ProductID int64 `json:"product_id"`
	UnitPrice int64 `json:"unit_price"`
	LineTotal int64 `json:"line_total"`
	Discount  int16 `json:"discount"`
	Quantity  int64 `json:"quantity"`
}

type SalesInvoiceTxParams struct {
	CashboxID int64 `json:"cashbox_id"`
	ShiftID   int64 `json:"shift_id"`

	Year            int32 `json:"year"`
	ClientID        int64 `json:"client_id"`
	InventoryID     int64 `json:"inventory_id"`
	Discount        int16 `json:"discount"`
	SubTotal        int64 `json:"sub_total"`
	DiscountedTotal int64 `json:"discounted_total"`
	GrandTotal      int64 `json:"grand_total"`

	Items []SalesInvoiceItem `json:"items"`

	CashboxAccountID int64         `json:"cashbox_account_id"`
	PriceListID      sql.NullInt64 `json:"price_list_id"`
	InvoiceTypeID    int64         `json:"invoice_type_id"`
}

type SalesInvoiceTxResult struct {
	Invoice Invoice               `json:"invoice"`
	Entry   Entry                 `json:"entry"`
	Balance ShiftsAccountsBalance `json:"balance"`
}

func (q *Queries) generateInvoiceIndex(ctx context.Context, cashboxID int64, year int32, indexType IndexType, invoiceTypeID int64) (int64, error) {
	arg := IncrementInvoicesIndexParams{
		CashboxID:     cashboxID,
		Year:          year,
		Type:          indexType,
		InvoiceTypeID: invoiceTypeID,
	}
	index, err := q.IncrementInvoicesIndex(ctx, arg)
	if err != nil {
		return 0, err
	}
	return index, nil
}

func (q *Queries) generateInvoiceNumber(ctx context.Context, referenceType EntryReferenceType, invoiceIndex, cashboxID, invoiceTypeID int64, year int32) (string, error) {
	var referenceCode string
	var err error

	cashbox, err := q.GetCashbox(ctx, cashboxID)
	if err != nil {
		return "", err
	}
	cashboxCode := cashbox.Code

	invoiceType, err := q.GetInvoiceType(ctx, invoiceTypeID)
	if err != nil {
		return "", err
	}

	switch referenceType {
	case EntryReferenceTypeSalesInvoice:
		referenceCode = "SA"
	case EntryReferenceTypeReturnInvoice:
		referenceCode = "RN"
	default:
		return "", errors.New("Invalid Reference Type")
	}

	return fmt.Sprintf("%s-%s-%s-%d-%05d", cashboxCode, invoiceType.Code, referenceCode, year, invoiceIndex), nil
}

func validateInvoiceAmounts(items []SalesInvoiceItem, discount int16, grandTotal, subTotal int64) error {
	if len(items) == 0 {
		return nil
	}

	var calculatedAmount, calculatedNetAmount int64

	for _, item := range items {
		itemPrice := item.UnitPrice * item.Quantity
		itemDiscountedPrice := itemPrice
		if item.Discount > 0 {
			itemDiscountedPrice = itemPrice - (itemPrice * int64(item.Discount) / 100)
		}
		calculatedAmount += itemPrice
		calculatedNetAmount += itemDiscountedPrice
	}

	calculatedAmount = calculatedAmount * (100 - int64(discount)) / 100

	if calculatedAmount != subTotal {
		return errors.New("Invalid Amount")
	}
	if calculatedNetAmount != grandTotal {
		return errors.New("Invalid Net Amount")
	}
	return nil
}

func (store *SQLStore) SalesInvoiceTx(ctx context.Context, arg SalesInvoiceTxParams) (SalesInvoiceTxResult, error) {
	var result SalesInvoiceTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		invoiceIndex, err := q.generateInvoiceIndex(ctx, arg.CashboxID, arg.Year, IndexTypeSales, arg.InvoiceTypeID)
		if err != nil {
			return err
		}

		invoiceCode, err := q.generateInvoiceNumber(ctx, EntryReferenceTypeSalesInvoice, invoiceIndex, arg.CashboxID, arg.InvoiceTypeID, arg.Year)
		if err != nil {
			return err
		}

		err = validateInvoiceAmounts(arg.Items, arg.Discount, arg.GrandTotal, arg.SubTotal)
		if err != nil {
			return err
		}

		invoice, err := q.CreateInvoice(ctx, CreateInvoiceParams{
			CashboxID:       arg.CashboxID,
			ShiftID:         arg.ShiftID,
			InvoiceCode:     invoiceCode,
			InvoiceIndex:    invoiceIndex,
			Year:            arg.Year,
			ClientID:        arg.ClientID,
			InventoryID:     arg.InventoryID,
			Discount:        arg.Discount,
			Subtotal:        arg.SubTotal,
			DiscountedTotal: arg.DiscountedTotal,
			GrandTotal:      arg.GrandTotal,
			InvoiceTypeID:   arg.InvoiceTypeID,
		})
		if err != nil {
			return err
		}

		_, err = q.CreateSalesInvoice(ctx, invoice.ID)
		if err != nil {
			return err
		}

		for _, item := range arg.Items {
			unitPrice := item.UnitPrice
			discount := item.Discount

			if arg.PriceListID.Valid {
				listPrice, priceErr := q.GetProductPriceFromList(ctx, GetProductPriceFromListParams{
					PriceListID: arg.PriceListID.Int64,
					ProductID:   item.ProductID,
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
				ProductID: item.ProductID,
				UnitPrice: unitPrice,
				LineTotal: lineTotal,
				Discount:  discount,
				Quantity:  item.Quantity,
			})
			if err != nil {
				return err
			}
		}

		entry, err := q.CreateEntryItem(ctx, CreateEntryItemParams{
			CashboxID:     arg.CashboxID,
			InventoryID:   arg.InventoryID,
			ReferenceType: EntryReferenceTypeSalesInvoice,
			ReferenceID:   invoice.ID,
			Amount:        arg.GrandTotal,
		})
		if err != nil {
			return err
		}

		addAccountBalanceArg := AddCashboxAccountBalanceParams{
			AccountID: arg.CashboxAccountID,
			ShiftID:   arg.ShiftID,
			Balance:   arg.GrandTotal,
		}
		balance, err := q.AddCashboxAccountBalance(ctx, addAccountBalanceArg)
		if err != nil {
			return err
		}

		for _, i := range arg.Items {
			addInventoryProductQuantityArg := AddInventoryProductQuantityParams{
				InventoryID: arg.InventoryID,
				ProductID:   i.ProductID,
				Quantity:    -i.Quantity,
			}
			err = q.AddInventoryProductQuantity(ctx, addInventoryProductQuantityArg)
			if err != nil {
				return err
			}
		}

		client, err := q.GetClient(ctx, invoice.ClientID)
		if err != nil {
			return err
		}
		// Wholesale clients are a central-office/B2B concept and don't earn
		// the retail loyalty program's points.
		if client.ClientType != "wholesale" {
			addPointsArg := AddClientLoyaltyPointsParams{
				ID:                 invoice.ClientID,
				TotalLoyaltyPoints: arg.GrandTotal,
				ValidLoyaltyPoints: arg.GrandTotal,
			}
			err = q.AddClientLoyaltyPoints(ctx, addPointsArg)
			if err != nil {
				return err
			}
		}

		result.Invoice = invoice
		result.Entry = entry
		result.Balance = balance

		return nil
	})

	return result, err
}
