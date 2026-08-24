package db

import (
	"context"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomExpense(t *testing.T) Expense {
	currency := createRandomCurrency(t)

	arg := CreateExpenseParams{
		Description:  util.RandomName(),
		Amount:       util.RandomAmount(),
		CurrencyCode: currency.Code,
	}

	expense, err := testQueries.CreateExpense(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, expense.Description, arg.Description)
	require.Equal(t, expense.Amount, arg.Amount)
	require.Equal(t, expense.CurrencyCode, arg.CurrencyCode)

	return expense
}

func TestCreateExpense(t *testing.T) {
	createRandomExpense(t)
}

func TestGetExpense(t *testing.T) {
	expense1 := createRandomExpense(t)

	expense2, err := testQueries.GetExpense(context.Background(), expense1.ID)
	require.NoError(t, err)
	require.Equal(t, expense1.ID, expense2.ID)
	require.Equal(t, expense1.Amount, expense2.Amount)
	require.Equal(t, expense1.Description, expense2.Description)
	require.Equal(t, expense1.CurrencyCode, expense2.CurrencyCode)

	require.WithinDuration(t, expense1.CreatedAt, expense2.CreatedAt, time.Second)
}

func TestListExpenses(t *testing.T) {
	for range 10 {
		createRandomExpense(t)
	}

	arg := ListExpensesParams{
		Limit:  5,
		Offset: 5,
	}

	expenses, err := testQueries.ListExpenses(context.Background(), arg)
	require.NoError(t, err)
	for _, expense := range expenses {
		require.NotEmpty(t, expense)
	}
}
