package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomInventory(t *testing.T) Inventory {
	arg := CreateInventoryParams{
		Name:      util.RandomName(),
		Type:      InventoryTypeStore,
		Code:      util.RandomCode(),
		Latitude:  util.RandomLongitudeLatitude(),
		Longitude: util.RandomLongitudeLatitude(),
	}

	inventory, err := testQueries.CreateInventory(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, inventory)
	require.Equal(t, arg.Name, inventory.Name)
	require.Equal(t, arg.Code, inventory.Code)
	require.Equal(t, arg.Latitude, inventory.Latitude)
	require.Equal(t, arg.Longitude, inventory.Longitude)

	require.NotZero(t, inventory.ID)
	require.NotZero(t, inventory.CreatedAt)

	return inventory
}

func TestCreateInventory(t *testing.T) {
	createRandomInventory(t)
}

func TestGetInventory(t *testing.T) {
	inventory1 := createRandomInventory(t)
	inventory2, err := testQueries.GetInventory(context.Background(), inventory1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, inventory2)

	require.Equal(t, inventory1.ID, inventory2.ID)
	require.Equal(t, inventory1.Code, inventory2.Code)
	require.Equal(t, inventory1.Latitude, inventory2.Latitude)
	require.Equal(t, inventory1.Longitude, inventory2.Longitude)
	require.WithinDuration(t, inventory1.CreatedAt, inventory2.CreatedAt, time.Second)
}

func TestDeleteInventory(t *testing.T) {
	inventory1 := createRandomInventory(t)
	err := testQueries.DeleteInventory(context.Background(), inventory1.ID)

	require.NoError(t, err)
	inventory2, err := testQueries.GetInventory(context.Background(), inventory1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, inventory2)
}

func TestListInventories(t *testing.T) {
	for range 10 {
		createRandomInventory(t)
	}
	arg := ListInventoriesParams{
		Limit:  5,
		Offset: 5,
	}

	inventories, err := testQueries.ListInventories(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, inventories, 5)
	for _, inventory := range inventories {
		require.NotEmpty(t, inventory)
	}
}

func TestUpdateInventory(t *testing.T) {
	inventory1 := createRandomInventory(t)
	arg := UpdateInventoryParams{
		ID:   inventory1.ID,
		Name: util.RandomName(),
	}

	err := testQueries.UpdateInventory(context.Background(), arg)
	require.NoError(t, err)

	inventory2, err := testQueries.GetInventory(context.Background(), inventory1.ID)

	require.Equal(t, inventory2.ID, arg.ID)
	require.Equal(t, inventory2.Name, arg.Name)
}

func createRandomVariant(t *testing.T) ProductVariant {
	product := createRandomProduct(t)
	variant, err := testQueries.CreateProductVariant(context.Background(), CreateProductVariantParams{
		ProductID: product.ID,
		Barcode:   util.RandomString(16),
		Price:     sql.NullInt64{Int64: util.RandomInt(100, 9999), Valid: true},
	})
	require.NoError(t, err)
	return variant
}

func TestAddInventoryStockQuantity(t *testing.T) {
	inventory := createRandomInventory(t)
	variant := createRandomVariant(t)

	first := util.RandomQuantity()
	row, err := testQueries.AddInventoryStockQuantity(context.Background(), AddInventoryStockQuantityParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
		Quantity:    first,
	})
	require.NoError(t, err)
	require.Equal(t, first, row.Quantity)

	extra := util.RandomQuantity()
	row, err = testQueries.AddInventoryStockQuantity(context.Background(), AddInventoryStockQuantityParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
		Quantity:    extra,
	})
	require.NoError(t, err)
	require.Equal(t, first+extra, row.Quantity)
}

func TestReceiveInventoryStockMovingAverage(t *testing.T) {
	inventory := createRandomInventory(t)
	variant := createRandomVariant(t)

	row, err := testQueries.ReceiveInventoryStock(context.Background(), ReceiveInventoryStockParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
		Quantity:    10,
		UnitCost:    100,
	})
	require.NoError(t, err)
	require.EqualValues(t, 10, row.Quantity)
	require.EqualValues(t, 100, row.AvgCost)

	// 10 @ 100 + 10 @ 200 => 20 @ 150
	row, err = testQueries.ReceiveInventoryStock(context.Background(), ReceiveInventoryStockParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
		Quantity:    10,
		UnitCost:    200,
	})
	require.NoError(t, err)
	require.EqualValues(t, 20, row.Quantity)
	require.EqualValues(t, 150, row.AvgCost)
}

func TestStockAdjustmentTx(t *testing.T) {
	inventory := createRandomInventory(t)
	variant := createRandomVariant(t)

	delta, err := testStore.StockAdjustmentTx(context.Background(), StockAdjustmentTxParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
		Mode:        StockAdjustmentModeDelta,
		Quantity:    7,
		Note:        "found a box",
	})
	require.NoError(t, err)
	require.EqualValues(t, 7, delta.OnHand)
	require.Equal(t, StockMovementReasonAdjustment, delta.Movement.Reason)

	count, err := testStore.StockAdjustmentTx(context.Background(), StockAdjustmentTxParams{
		InventoryID: inventory.ID,
		VariantID:   variant.ID,
		Mode:        StockAdjustmentModeCount,
		Quantity:    3,
		Note:        "cycle count",
	})
	require.NoError(t, err)
	require.EqualValues(t, 3, count.OnHand)
	require.EqualValues(t, -4, count.Movement.Quantity)
	require.Equal(t, StockMovementReasonCount, count.Movement.Reason)

	moves, err := testQueries.ListStockMovements(context.Background(), ListStockMovementsParams{
		InventoryID: sql.NullInt64{Int64: inventory.ID, Valid: true},
		VariantID:   sql.NullInt64{Int64: variant.ID, Valid: true},
		PageLimit:   50,
	})
	require.NoError(t, err)
	require.Len(t, moves, 2)
}
