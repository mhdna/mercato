package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomPurchase(t *testing.T) Purchase {
	supplier := createRandomSupplier(t)
	currency := createRandomCurrency(t)
	arg := CreatePurchaseParams{
		SupplierID:   supplier.ID,
		CurrencyCode: currency.Code,
		Code:         util.RandomCode(),
		PurchasedAt:  time.Now(),
	}

	purchase, err := testQueries.CreatePurchase(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, purchase)
	require.Equal(t, arg.SupplierID, purchase.SupplierID)
	require.Equal(t, PurchaseStatusDraft, purchase.Status)
	require.WithinDuration(t, arg.PurchasedAt, purchase.PurchasedAt, time.Second)

	require.NotZero(t, purchase.ID)
	return purchase
}

func TestCreatePurchase(t *testing.T) {
	createRandomPurchase(t)
}

func TestGetPurchase(t *testing.T) {
	purchase1 := createRandomPurchase(t)
	purchase2, err := testQueries.GetPurchase(context.Background(), purchase1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, purchase2)

	require.Equal(t, purchase1.ID, purchase2.ID)
	require.Equal(t, purchase1.SupplierID, purchase2.SupplierID)
	require.WithinDuration(t, purchase1.PurchasedAt, purchase2.PurchasedAt, time.Second)
}

func TestListPurchases(t *testing.T) {
	for range 10 {
		createRandomPurchase(t)
	}

	arg := ListPurchasesParams{
		Search:     "",
		PageSize:   5,
		PageOffset: 5,
	}

	purchases, err := testQueries.ListPurchases(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, purchases, 5)

	for _, purchase := range purchases {
		require.NotEmpty(t, purchase)
	}
}

func TestAddPurchaseItem(t *testing.T) {
	variant := createRandomVariant(t)
	purchase := createRandomPurchase(t)
	currency := createRandomCurrency(t)
	arg := AddPurchaseItemParams{
		PurchaseID:   sql.NullInt64{Int64: purchase.ID, Valid: true},
		VariantID:    sql.NullInt64{Int64: variant.ID, Valid: true},
		Quantity:     util.RandomQuantity(),
		UnitPrice:    util.RandomAmount(),
		CurrencyCode: currency.Code,
	}

	purchased_item, err := testQueries.AddPurchaseItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, purchase)
	require.Equal(t, arg.PurchaseID, purchased_item.PurchaseID)
	require.Equal(t, arg.VariantID, purchased_item.VariantID)
	require.Equal(t, arg.Quantity, purchased_item.Quantity)
	require.Equal(t, arg.UnitPrice, purchased_item.UnitPrice)
	require.Equal(t, arg.CurrencyCode, purchased_item.CurrencyCode)

	require.NotZero(t, purchased_item.ID)
}

func TestPurchaseReceiveTx(t *testing.T) {
	inventory := createRandomInventory(t)
	variant := createRandomVariant(t)
	supplier := createRandomSupplier(t)
	currency := createRandomCurrency(t)

	created, err := testStore.CreatePurchaseTx(context.Background(), CreatePurchaseTxParams{
		SupplierID:   supplier.ID,
		InventoryID:  inventory.ID,
		CurrencyCode: currency.Code,
		Items: []PurchaseItemParams{
			{VariantID: sql.NullInt64{Int64: variant.ID, Valid: true}, Quantity: 5, UnitPrice: 200},
		},
	})
	require.NoError(t, err)
	require.EqualValues(t, 1000, created.Purchase.GrandTotal)

	received, err := testStore.PurchaseReceiveTx(context.Background(), PurchaseReceiveTxParams{
		PurchaseID: created.Purchase.ID,
	})
	require.NoError(t, err)
	require.Equal(t, PurchaseStatusReceived, received.Purchase.Status)
	require.Len(t, received.Movements, 1)

	stock, err := testQueries.GetInventoryStock(context.Background(), GetInventoryStockParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, 5, stock.Quantity)
	require.EqualValues(t, 200, stock.AvgCost)

	// Receiving again is a no-op (idempotent on status).
	again, err := testStore.PurchaseReceiveTx(context.Background(), PurchaseReceiveTxParams{
		PurchaseID: created.Purchase.ID,
	})
	require.NoError(t, err)
	require.Equal(t, PurchaseStatusReceived, again.Purchase.Status)
	require.Empty(t, again.Movements)
}

func TestAddPurchasedProduct(t *testing.T) {
	product := createRandomProduct(t)
	supplier := createRandomSupplier(t)

	arg := AddPurchasedProductParams{
		ProductID:  product.ID,
		SupplierID: supplier.ID,
	}

	result, err := testQueries.AddPurchasedProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, arg.ProductID, result.ProductID)
	require.Equal(t, arg.SupplierID, result.SupplierID)
}

func TestAddPurchasedProductCost(t *testing.T) {
	product := createRandomProduct(t)
	supplier := createRandomSupplier(t)
	currency := createRandomCurrency(t)

	ps, err := testQueries.AddPurchasedProduct(context.Background(), AddPurchasedProductParams{
		ProductID:  product.ID,
		SupplierID: supplier.ID,
	})
	require.NoError(t, err)

	cost := util.RandomAmount()
	arg := AddPurchasedProductCostParams{
		ProductSupplierID: ps.ID,
		UnitCost:          cost,
		CurrencyCode:      currency.Code,
	}

	result, err := testQueries.AddPurchasedProductCost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, cost, result.UnitCost)
	require.Equal(t, currency.Code, result.CurrencyCode)
}
