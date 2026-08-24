package db

import (
	"context"
	"strconv"
	"testing"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func TestCreateProductTxGeneratesUniqueBarcodeWhenNoneGiven(t *testing.T) {
	store := NewStore(testDB)

	result, err := store.CreateProductTx(context.Background(), CreateProductTxParams{
		Code:        util.RandomString(8),
		Name:        util.RandomString(20),
		Description: util.RandomString(50),
		Price:       util.RandomAmount(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Product)
	require.NotEmpty(t, result.Variant)
	require.Equal(t, result.Product.ID, result.Variant.ProductID)
	require.NotEmpty(t, result.Variant.Barcode)
	require.True(t, result.Variant.Price.Valid)

	// The variant must actually be retrievable by the barcode that came back
	// — proves it was written to product_variants, not silently dropped.
	fetched, err := testQueries.GetProductVariantByBarcode(context.Background(), result.Variant.Barcode)
	require.NoError(t, err)
	require.Equal(t, result.Variant.ID, fetched.ID)
}

func TestCreateProductTxUsesGivenBarcode(t *testing.T) {
	store := NewStore(testDB)

	barcode := util.RandomInt(1_000_000_000, 9_000_000_000)
	result, err := store.CreateProductTx(context.Background(), CreateProductTxParams{
		Code:        util.RandomString(8),
		Name:        util.RandomString(20),
		Description: util.RandomString(50),
		Barcode:     &barcode,
	})
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(barcode, 10), result.Variant.Barcode)
}

func TestCreateProductTxAssignsColorAndSize(t *testing.T) {
	store := NewStore(testDB)

	color, err := testQueries.CreateColor(context.Background(), CreateColorParams{
		Name:     util.RandomString(10),
		HexValue: "#" + util.RandomString(6),
	})
	require.NoError(t, err)

	result, err := store.CreateProductTx(context.Background(), CreateProductTxParams{
		Code:        util.RandomString(8),
		Name:        util.RandomString(20),
		Description: util.RandomString(50),
		ColorID:     &color.ID,
	})
	require.NoError(t, err)
	require.True(t, result.Variant.ColorID.Valid)
	require.Equal(t, color.ID, result.Variant.ColorID.Int64)
	require.False(t, result.Variant.SizeID.Valid)
}
