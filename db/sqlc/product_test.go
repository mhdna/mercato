package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

// TODO: add price list and discount list tests
func createRandomProduct(t *testing.T) Product {
	arg := CreateProductParams{
		Name:        util.RandomString(20),
		Code:        util.RandomString(8),
		Description: util.RandomString(200),
	}

	product, err := testQueries.CreateProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, product)
	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.Code, product.Code)
	require.Equal(t, arg.Description, product.Description)

	require.NotZero(t, product.ID)
	require.NotZero(t, product.CreatedAt)

	return product
}

func TestCreateProduct(t *testing.T) {
	createRandomProduct(t)
}

func TestGetProduct(t *testing.T) {
	product1 := createRandomProduct(t)
	product2, err := testQueries.GetProduct(context.Background(), product1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Code, product2.Code)
	require.Equal(t, product1.Description, product2.Description)
	require.WithinDuration(t, product1.CreatedAt, product2.CreatedAt, time.Second)
}

func TestListProducts(t *testing.T) {
	for range 10 {
		createRandomProduct(t)
	}

	arg := ListProductsParams{
		PageLimit:         5,
		PageOffset:        5,
		AttributeValueIds: []int64{},
	}

	products, err := testQueries.ListProducts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, products, 5)
	for _, product := range products {
		require.NotEmpty(t, product)
	}
}

func TestUpdateProduct(t *testing.T) {
	product := createRandomProduct(t)

	arg := UpdateProductParams{
		ID:          product.ID,
		Name:        util.RandomString(20),
		Code:        util.RandomString(8),
		Description: util.RandomString(200),
		IsActive:    true,
	}

	err := testQueries.UpdateProduct(context.Background(), arg)
	require.NoError(t, err)

	updated, err := testQueries.GetProduct(context.Background(), product.ID)
	require.NoError(t, err)
	require.Equal(t, arg.Name, updated.Name)
	require.Equal(t, arg.Code, updated.Code)
	require.Equal(t, arg.Description, updated.Description)
}

func TestDeleteProduct(t *testing.T) {
	product := createRandomProduct(t)

	err := testQueries.DeleteProduct(context.Background(), product.ID)
	require.NoError(t, err)

	_, err = testQueries.GetProduct(context.Background(), product.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
}
