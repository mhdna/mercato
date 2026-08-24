package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomCurrency(t *testing.T) Currency {
	arg := CreateCurrencyParams{
		Name:                   util.RandomName(),
		Code:                   util.RandomCode(),
		Symbol:                 util.RandomCode(),
		ValueInDefaultCurrency: util.RandomAmount(),
		UnitsPerUsdMicros:      1_000_000,
		CashRoundingUnit:       1,
	}

	currency, err := testQueries.CreateCurrency(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, currency)
	require.Equal(t, arg.Name, currency.Name)
	require.Equal(t, arg.Code, currency.Code)
	require.Equal(t, arg.Symbol, currency.Symbol)
	require.Equal(t, arg.ValueInDefaultCurrency, currency.ValueInDefaultCurrency)

	return currency
}

func TestCreateCurrency(t *testing.T) {
	createRandomCurrency(t)
}

func TestGetCurrency(t *testing.T) {
	currency1 := createRandomCurrency(t)
	currency2, err := testQueries.GetCurrency(context.Background(), currency1.Code)

	require.NoError(t, err)
	require.NotEmpty(t, currency2)

	require.Equal(t, currency1.Name, currency2.Name)
	require.Equal(t, currency1.Code, currency2.Code)
	require.Equal(t, currency1.Symbol, currency2.Symbol)
	require.Equal(t, currency1.ValueInDefaultCurrency, currency2.ValueInDefaultCurrency)
}

func TestDeleteCurrency(t *testing.T) {
	currency1 := createRandomCurrency(t)
	err := testQueries.DeleteCurrency(context.Background(), currency1.Code)
	require.NoError(t, err)

	currency2, err := testQueries.GetCurrency(context.Background(), currency1.Code)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, currency2)
}

func TestListCurrencies(t *testing.T) {
	for range 10 {
		createRandomCurrency(t)
	}
	arg := ListCurrenciesParams{
		Limit:  5,
		Offset: 5,
	}
	currencies, err := testQueries.ListCurrencies(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, currencies, 5)
	for _, currency := range currencies {
		require.NotEmpty(t, currency)
	}
}
