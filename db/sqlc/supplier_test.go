package db

import (
	"context"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomSupplier(t *testing.T) Supplier {
	arg := CreateSupplierParams{
		Name:             util.RandomName(),
		Phone:            util.RandomPhone(),
		Country:          util.RandomString(20),
		Address:          util.RandomString(100),
		AddressLatitude:  util.RandomLongitudeLatitude(),
		AddressLongitude: util.RandomLongitudeLatitude(),
	}

	supplier, err := testQueries.CreateSupplier(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, supplier)
	require.Equal(t, arg.Name, supplier.Name)
	require.Equal(t, arg.Phone, supplier.Phone)
	require.Equal(t, arg.Country, supplier.Country)
	require.Equal(t, arg.Address, supplier.Address)
	require.Equal(t, arg.AddressLatitude, supplier.AddressLatitude)
	require.Equal(t, arg.AddressLongitude, supplier.AddressLongitude)
	require.NotZero(t, supplier.ID)
	require.NotZero(t, supplier.CreatedAt)

	return supplier
}

func TestCreateSupplier(t *testing.T) {
	createRandomSupplier(t)
}

func TestGetSupplier(t *testing.T) {
	supplier1 := createRandomSupplier(t)
	supplier2, err := testQueries.GetSupplier(context.Background(), supplier1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, supplier2)

	require.Equal(t, supplier1.ID, supplier2.ID)
	require.Equal(t, supplier1.Name, supplier2.Name)
	require.Equal(t, supplier1.Phone, supplier2.Phone)
	require.Equal(t, supplier1.Country, supplier2.Country)
	require.Equal(t, supplier1.Address, supplier2.Address)
	require.Equal(t, supplier1.AddressLatitude, supplier2.AddressLatitude)
	require.Equal(t, supplier1.AddressLongitude, supplier2.AddressLongitude)
	require.WithinDuration(t, supplier1.CreatedAt, supplier2.CreatedAt, time.Second)
}

func TestListSuppliers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomSupplier(t)
	}

	arg := ListSuppliersParams{
		Search:     "",
		PageSize:   5,
		PageOffset: 5,
	}

	suppliers, err := testQueries.ListSuppliers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, suppliers, 5)

	for _, supplier := range suppliers {
		require.NotEmpty(t, supplier)
	}
}
