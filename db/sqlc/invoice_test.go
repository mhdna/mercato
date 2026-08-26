package db

// TODO: Test line total

import (
	"context"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomSalesInvoice(t *testing.T) Invoice {
	client := createRandomClient(t)
	cashbox := createRandomCashbox(t)
	shift := createRandomShift(t)
	inventory := createRandomInventory(t)
	invoiceType, err := testQueries.GetDefaultInvoiceType(context.Background())
	require.NoError(t, err)
	discount := int16(0)
	grandTotal := util.RandomAmount()
	subTotal := util.RandomAmount()
	discountedTotal := util.RandomAmount()

	arg := CreateInvoiceParams{
		InventoryID:     inventory.ID,
		ClientID:        client.ID,
		CashboxID:       cashbox.ID,
		ShiftID:         shift.ID,
		Discount:        discount,
		Subtotal:        subTotal,
		GrandTotal:      grandTotal,
		DiscountedTotal: discountedTotal,
		InvoiceCode:     "TEST-SA-2026-00001",
		InvoiceIndex:    1,
		Year:            int32(time.Now().Year()),
		InvoiceTypeID:   invoiceType.ID,
	}

	invoice, err := testQueries.CreateInvoice(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, invoice)

	_, err = testQueries.CreateSalesInvoice(context.Background(), invoice.ID)
	require.NoError(t, err)

	require.Equal(t, arg.InvoiceIndex, invoice.InvoiceIndex)
	require.Equal(t, arg.InvoiceCode, invoice.InvoiceCode)
	require.Equal(t, arg.Year, invoice.Year)
	require.Equal(t, arg.InventoryID, invoice.InventoryID)
	require.Equal(t, arg.ClientID, invoice.ClientID)
	require.Equal(t, arg.Discount, invoice.Discount)
	require.Equal(t, arg.GrandTotal, invoice.GrandTotal)
	require.Equal(t, arg.Subtotal, invoice.Subtotal)
	require.Equal(t, arg.DiscountedTotal, invoice.DiscountedTotal)

	require.NotZero(t, invoice.ID)
	require.NotZero(t, invoice.CreatedAt)

	return invoice
}

func TestCreateSalesInvoice(t *testing.T) {
	createRandomSalesInvoice(t)
}

func TestAddInvoiceProduct(t *testing.T) {
	salesInvoice := createRandomSalesInvoice(t)
	product := createRandomProduct(t)

	arg := AddInvoiceProductParams{
		InvoiceID: salesInvoice.ID,
		ProductID: product.ID,
		Quantity:  util.RandomQuantity(),
	}

	result, err := testQueries.AddInvoiceProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, arg.InvoiceID, result.InvoiceID)
	require.Equal(t, arg.ProductID, result.ProductID)
	require.Equal(t, arg.UnitPrice, result.UnitPrice)
	require.Equal(t, arg.Quantity, result.Quantity)
}

func TestListInvoices(t *testing.T) {
	for range 10 {
		createRandomSalesInvoice(t)
	}

	arg := ListInvoicesParams{
		Limit:  5,
		Offset: 5,
	}

	invoices, err := testQueries.ListInvoices(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, invoices, 5)
	for _, invoice := range invoices {
		require.NotEmpty(t, invoice)
	}
}

func TestIncrementInvoicesIndex(t *testing.T) {
	cashbox := createRandomCashbox(t)
	year := int32(time.Now().Year())
	invoiceType, err := testQueries.GetDefaultInvoiceType(context.Background())
	require.NoError(t, err)

	arg := IncrementInvoicesIndexParams{
		Year:          year,
		CashboxID:     cashbox.ID,
		Type:          IndexTypeSales,
		InvoiceTypeID: invoiceType.ID,
	}

	index1, err := testQueries.IncrementInvoicesIndex(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), index1)

	index2, err := testQueries.IncrementInvoicesIndex(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(2), index2)

	index3, err := testQueries.IncrementInvoicesIndex(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(3), index3)
}

func createRandomReturnInvoice(t *testing.T) Invoice {
	salesInvoice := createRandomSalesInvoice(t)
	cashbox := createRandomCashbox(t)
	shift := createRandomShift(t)
	client := createRandomClient(t)
	inventory := createRandomInventory(t)
	invoiceType, err := testQueries.GetDefaultInvoiceType(context.Background())
	require.NoError(t, err)

	arg := CreateInvoiceParams{
		CashboxID:       cashbox.ID,
		ShiftID:         shift.ID,
		InvoiceCode:     "TEST-RN-2026-00001",
		InvoiceIndex:    1,
		Year:            int32(time.Now().Year()),
		ClientID:        client.ID,
		InventoryID:     inventory.ID,
		Discount:        0,
		Subtotal:        util.RandomAmount(),
		DiscountedTotal: util.RandomAmount(),
		GrandTotal:      util.RandomAmount(),
		InvoiceTypeID:   invoiceType.ID,
	}

	invoice, err := testQueries.CreateInvoice(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, invoice)

	returnInvoice, err := testQueries.CreateReturnInvoice(context.Background(), CreateReturnInvoiceParams{
		InvoiceID:      invoice.ID,
		SalesInvoiceID: salesInvoice.ID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, returnInvoice)
	require.Equal(t, invoice.ID, returnInvoice.InvoiceID)
	require.Equal(t, salesInvoice.ID, returnInvoice.SalesInvoiceID)

	require.NotZero(t, invoice.ID)
	require.NotZero(t, invoice.CreatedAt)

	return invoice
}

func TestCreateReturnInvoice(t *testing.T) {
	createRandomReturnInvoice(t)
}

func TestGetReturnInvoice(t *testing.T) {
	invoice := createRandomReturnInvoice(t)
	returnInvoice, err := testQueries.GetReturnInvoice(context.Background(), invoice.ID)
	require.NoError(t, err)
	require.Equal(t, invoice.ID, returnInvoice.InvoiceID)
}

func TestAddReturnInvoiceProduct(t *testing.T) {
	invoice := createRandomReturnInvoice(t)
	product := createRandomProduct(t)

	arg := AddInvoiceProductParams{
		InvoiceID: invoice.ID,
		ProductID: product.ID,
		Quantity:  util.RandomQuantity(),
	}

	result, err := testQueries.AddInvoiceProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, arg.InvoiceID, result.InvoiceID)
	require.Equal(t, arg.ProductID, result.ProductID)
	require.Equal(t, arg.UnitPrice, result.UnitPrice)
	require.Equal(t, arg.Quantity, result.Quantity)
}

func TestDecrementInvoicesIndex(t *testing.T) {
	cashbox := createRandomCashbox(t)
	year := int32(time.Now().Year())
	invoiceType, err := testQueries.GetDefaultInvoiceType(context.Background())
	require.NoError(t, err)

	arg := IncrementInvoicesIndexParams{
		Year:          year,
		CashboxID:     cashbox.ID,
		Type:          IndexTypeReturn,
		InvoiceTypeID: invoiceType.ID,
	}

	index1, err := testQueries.IncrementInvoicesIndex(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), index1)
}
