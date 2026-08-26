package db

import (
	"context"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

// func TestTransferTX(t *testing.T) {
// 	store := NewStore(testDB)

// 	inventory1 := createRandomInventory(t)
// 	inventory2 := createRandomInventory(t)

// 	errs := make(chan error)
// 	results := make(chan TransferTxResult)

// 	for range 5 {
// 		go func() {
// 			result, err := store.TransferTx(context.Background(), TransferTxParams{
// 				FromInventoryID: inventory1.ID,
// 				ToInventoryID:   inventory2.ID,
// 				TransferType:    TransferTypeProducts,
// 			})
// 			errs <- err
// 			results <- result
// 		}()
// 	}

// 	for range 5 {
// 		err := <-errs
// 		require.NoError(t, err)

// 		result := <-results
// 		require.NotEmpty(t, results)

// 		// check transfer
// 		transfer := result.Transfer

// 		items := result.Items

// 		require.NotEmpty(t, transfer)
// 		require.Equal(t, inventory1.ID, transfer.FromInventoryID)
// 		require.Equal(t, inventory2.ID, transfer.ToInventoryID)
// 		require.NotZero(t, transfer.ID)
// 		require.NotZero(t, transfer.CreatedAt)

// 		for _, i := range items {
// 			require.NotZero(t, i.Quantity)
// 		}
// 	}
// }

// TODO add tests to see if invoices are incrementing properly
// TODO add tests to cover different cashboxes, accounts, etc.
func TestSalesInvoiceTx(t *testing.T) {
	store := NewStore(testDB)
	discount := util.RandomDiscount()
	cashbox := createRandomCashbox(t)
	account := createRandomAccount(t)
	inventory := createRandomInventory(t)
	shift := createRandomShift(t)
	client := createRandomClient(t)
	invoiceType, err := testQueries.GetDefaultInvoiceType(context.Background())
	require.NoError(t, err)
	grandTotal := util.RandomAmount()
	subTotal := util.RandomAmount()
	discountedTotal := util.RandomAmount()

	n := int64(5)

	errs := make(chan error)
	results := make(chan SalesInvoiceTxResult)

	for range n {
		go func() {
			res, err := store.SalesInvoiceTx(context.Background(), SalesInvoiceTxParams{
				CashboxID:        cashbox.ID,
				CashboxAccountID: account.ID,
				ShiftID:          shift.ID,
				InventoryID:      inventory.ID,
				ClientID:         client.ID,
				Discount:         discount,
				GrandTotal:       grandTotal,
				SubTotal:         subTotal,
				DiscountedTotal:  discountedTotal,
				Year:             int32(time.Now().Year()),
				InvoiceTypeID:    invoiceType.ID,
			})
			errs <- err
			results <- res
		}()
	}

	for i := int64(1); i <= n; i++ {
		err := <-errs
		res := <-results
		require.NoError(t, err)

		salesInvoice := res.Invoice
		require.NotEmpty(t, salesInvoice)
		require.Equal(t, inventory.ID, salesInvoice.InventoryID)
		require.Equal(t, cashbox.ID, salesInvoice.CashboxID)
		require.Equal(t, client.ID, salesInvoice.ClientID)
		require.Equal(t, discount, salesInvoice.Discount)
		require.Equal(t, grandTotal, salesInvoice.GrandTotal)
		require.Equal(t, subTotal, salesInvoice.Subtotal)
		require.Equal(t, discountedTotal, salesInvoice.DiscountedTotal)
		require.NotZero(t, salesInvoice.ID)
		require.NotZero(t, salesInvoice.CreatedAt)

		entry := res.Entry
		require.Equal(t, cashbox.ID, entry.CashboxID)
		require.Equal(t, inventory.ID, entry.InventoryID)
		// TODO fix this
		// netAmount, err := util.CalculateNetAmount(amount, discount)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		require.Equal(t, grandTotal, entry.Amount)
		require.Equal(t, salesInvoice.ID, entry.ReferenceID)
		require.Equal(t, EntryReferenceTypeSalesInvoice, entry.ReferenceType)
		require.NotZero(t, entry.CreatedAt)
		require.NotZero(t, entry.ID)

		// // check accounts and their balances
		// resAccount := res.Account
		// require.Equal(t, account.ID, resAccount.ID)
		// iterationNetAmount := netAmount * i
		// iterationAmount := amount * i
		// fmt.Printf(">> tx: %d + %d (%d - %d%%) = %d\n", account.Balance, iterationNetAmount, iterationAmount, discount, resAccount.Balance)
		// diff := resAccount.Balance - account.Balance
		// require.Equal(t, iterationNetAmount, diff)

		// TODO: check this (might cause problems dividing by 0)
		// require.True(t, diff%iterationNetAmount == 0)

		// // check shift Balance
		// resShift := res.Shift
		// require.Equal(t, shift.ID, resShift.ID)
		// require.Equal(t, shift.TotalBalance+iterationNetAmount, resShift.TotalBalance)
		// require.Equal(t, shift.CashboxID, resShift.CashboxID)
		// require.Equal(t, shift.TotalOpeningBalance, resShift.TotalOpeningBalance)
	}
}
