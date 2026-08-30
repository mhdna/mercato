package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomPriceList(t *testing.T) PriceList {
	arg := CreatePriceListParams{
		Name:      "Test Price List",
		IsActive:  true,
		IsDefault: false,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	}
	priceList, err := testQueries.CreatePriceList(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, priceList)
	return priceList
}

func TestCreatePriceList(t *testing.T) {
	createRandomPriceList(t)
}

func TestGetPriceList(t *testing.T) {
	priceList := createRandomPriceList(t)
	got, err := testQueries.GetPriceList(context.Background(), priceList.ID)
	require.NoError(t, err)
	require.Equal(t, priceList.ID, got.ID)
	require.Equal(t, priceList.Name, got.Name)
}

func TestListPriceLists(t *testing.T) {
	for range 5 {
		createRandomPriceList(t)
	}
	arg := ListPriceListsParams{Search: "", PageSize: 5, PageOffset: 0}
	list, err := testQueries.ListPriceLists(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, list, 5)
}

func TestCreatePriceListItem(t *testing.T) {
	priceList := createRandomPriceList(t)
	product := createRandomProduct(t)

	arg := CreatePriceListItemParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
		Price:       999,
	}
	item, err := testQueries.CreatePriceListItem(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, priceList.ID, item.PriceListID)
	require.Equal(t, product.ID, item.ProductID)
	require.Equal(t, int64(999), item.Price)
}

func TestGetProductPriceFromList(t *testing.T) {
	priceList := createRandomPriceList(t)
	product := createRandomProduct(t)

	_, err := testQueries.CreatePriceListItem(context.Background(), CreatePriceListItemParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
		Price:       500,
	})
	require.NoError(t, err)

	got, err := testQueries.GetProductPriceFromList(context.Background(), GetProductPriceFromListParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
	})
	require.NoError(t, err)
	require.Equal(t, int64(500), got.Price)
}

func TestUpdatePriceListItem(t *testing.T) {
	priceList := createRandomPriceList(t)
	product := createRandomProduct(t)

	_, err := testQueries.CreatePriceListItem(context.Background(), CreatePriceListItemParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
		Price:       100,
	})
	require.NoError(t, err)

	err = testQueries.UpdatePriceListItem(context.Background(), UpdatePriceListItemParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
		Price:       200,
	})
	require.NoError(t, err)

	got, err := testQueries.GetProductPriceFromList(context.Background(), GetProductPriceFromListParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
	})
	require.NoError(t, err)
	require.Equal(t, int64(200), got.Price)
}

func TestDeletePriceListItem(t *testing.T) {
	priceList := createRandomPriceList(t)
	product := createRandomProduct(t)

	_, err := testQueries.CreatePriceListItem(context.Background(), CreatePriceListItemParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
		Price:       300,
	})
	require.NoError(t, err)

	err = testQueries.DeletePriceListItem(context.Background(), DeletePriceListItemParams{
		PriceListID: priceList.ID,
		ProductID:   product.ID,
	})
	require.NoError(t, err)
}

func TestListPriceListItems(t *testing.T) {
	priceList := createRandomPriceList(t)
	for range 3 {
		product := createRandomProduct(t)
		_, err := testQueries.CreatePriceListItem(context.Background(), CreatePriceListItemParams{
			PriceListID: priceList.ID,
			ProductID:   product.ID,
			Price:       100,
		})
		require.NoError(t, err)
	}

	items, err := testQueries.ListPriceListItems(context.Background(), priceList.ID)
	require.NoError(t, err)
	require.Len(t, items, 3)
}

func TestCreateDiscountList(t *testing.T) {
	arg := CreateDiscountListParams{
		Name:      "Test Discount List",
		IsActive:  true,
		IsDefault: false,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	}
	discountList, err := testQueries.CreateDiscountList(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, discountList)
}

func TestGetDiscountList(t *testing.T) {
	arg := CreateDiscountListParams{
		Name:      "Get Test DL",
		IsActive:  true,
		IsDefault: false,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	}
	dl, err := testQueries.CreateDiscountList(context.Background(), arg)
	require.NoError(t, err)

	got, err := testQueries.GetDiscountList(context.Background(), dl.ID)
	require.NoError(t, err)
	require.Equal(t, dl.ID, got.ID)
}

func TestCreateDiscountListItem(t *testing.T) {
	dl, err := testQueries.CreateDiscountList(context.Background(), CreateDiscountListParams{
		Name:      "DL for items",
		IsActive:  true,
		IsDefault: false,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	product := createRandomProduct(t)
	item, err := testQueries.CreateDiscountListItem(context.Background(), CreateDiscountListItemParams{
		DiscountListID: dl.ID,
		ProductID:      product.ID,
		Discount:       15,
	})
	require.NoError(t, err)
	require.Equal(t, int16(15), item.Discount)
}

func TestGetProductDiscountFromList(t *testing.T) {
	dl, err := testQueries.CreateDiscountList(context.Background(), CreateDiscountListParams{
		Name:      "DL for get",
		IsActive:  true,
		IsDefault: false,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	product := createRandomProduct(t)
	_, err = testQueries.CreateDiscountListItem(context.Background(), CreateDiscountListItemParams{
		DiscountListID: dl.ID,
		ProductID:      product.ID,
		Discount:       25,
	})
	require.NoError(t, err)

	got, err := testQueries.GetProductDiscountFromList(context.Background(), GetProductDiscountFromListParams{
		DiscountListID: dl.ID,
		ProductID:      product.ID,
	})
	require.NoError(t, err)
	require.Equal(t, int16(25), got.Discount)
}

func TestDeleteDiscountListItem(t *testing.T) {
	dl, err := testQueries.CreateDiscountList(context.Background(), CreateDiscountListParams{
		Name:      "DL for delete",
		IsActive:  true,
		IsDefault: false,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	product := createRandomProduct(t)
	_, err = testQueries.CreateDiscountListItem(context.Background(), CreateDiscountListItemParams{
		DiscountListID: dl.ID,
		ProductID:      product.ID,
		Discount:       10,
	})
	require.NoError(t, err)

	err = testQueries.DeleteDiscountListItem(context.Background(), DeleteDiscountListItemParams{
		DiscountListID: dl.ID,
		ProductID:      product.ID,
	})
	require.NoError(t, err)
}
