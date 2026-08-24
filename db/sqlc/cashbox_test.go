package db

import (
	"context"
	"testing"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomCashbox(t *testing.T) Cashbox {
	arg := CreateCashboxParams{
		Code:     util.RandomCode(),
		Name:     util.RandomName(),
		IsActive: true,
	}

	cashbox, err := testQueries.CreateCashbox(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, cashbox)
	require.Equal(t, cashbox.Code, arg.Code)
	require.Equal(t, cashbox.Name, arg.Name)
	require.Equal(t, cashbox.IsActive, arg.IsActive)

	return cashbox
}

func TestGetCashbox(t *testing.T) {
	cashbox1 := createRandomCashbox(t)

	cashbox2, err := testQueries.GetCashbox(context.Background(), cashbox1.ID)

	require.NoError(t, err)
	require.Equal(t, cashbox1.ID, cashbox2.ID)
	require.Equal(t, cashbox1.Code, cashbox2.Code)
	require.Equal(t, cashbox1.Name, cashbox2.Name)
	require.Equal(t, cashbox1.IsActive, cashbox2.IsActive)
}

func TestCreateCashbox(t *testing.T) {
	createRandomCashbox(t)
}

func TestListCashboxes(t *testing.T) {
	for range 10 {
		createRandomCashbox(t)
	}

	limit := 5
	offset := 0

	arg := ListCashboxesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	cashboxes, err := testQueries.ListCashboxes(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, cashboxes, limit)
	for _, cashbox := range cashboxes {
		// TODO: complete this
		require.NotEmpty(t, cashbox)
	}
}

func TestUpdateCashbox(t *testing.T) {
	cashbox := createRandomCashbox(t)

	arg := UpdateCashboxParams{
		ID:       cashbox.ID,
		Code:     util.RandomCode(),
		Name:     util.RandomName(),
		IsActive: !cashbox.IsActive,
	}

	updated, err := testQueries.UpdateCashbox(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, arg.ID, updated.ID)
	require.Equal(t, arg.Code, updated.Code)
	require.Equal(t, arg.Name, updated.Name)
	require.Equal(t, arg.IsActive, updated.IsActive)
}

func TestUpdateCashboxAccount(t *testing.T) {
	account := createRandomAccount(t)

	arg := UpdateCashboxAccountParams{
		ID:           account.ID,
		Name:         util.RandomName(),
		CurrencyCode: "USD",
	}

	updated, err := testQueries.UpdateCashboxAccount(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, arg.ID, updated.ID)
	require.Equal(t, arg.Name, updated.Name)
}
