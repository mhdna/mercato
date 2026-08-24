package db

import (
	"context"
	"testing"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) CashboxAccount {

	// shift := createRandomShift(t)
	// balance := util.RandomAmount()
	cashboxAccountName := util.RandomName()

	// arg := CreateCashboxAccountParams{
	// 	Type:           util.RandomName(),
	// 	ShiftID:        shift.ID,
	// 	OpeningBalance: balance,
	// 	Balance:        balance,
	// }

	account, err := testQueries.CreateCashboxAccount(context.Background(), CreateCashboxAccountParams{
		Name:         cashboxAccountName,
		CurrencyCode: "USD",
	})
	require.NoError(t, err)
	require.Equal(t, account.Name, cashboxAccountName)
	// require.Equal(t, account.Type, arg.Type)
	// require.Equal(t, account.ShiftID, arg.ShiftID)
	// require.Equal(t, account.OpeningBalance, arg.OpeningBalance)
	// require.Equal(t, account.Balance, arg.Balance)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account1 := createRandomAccount(t)

	account2, err := testQueries.GetCashboxAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Name, account2.Name)
	// require.Equal(t, account1.Type, account2.Type)
	// require.Equal(t, account1.ShiftID, account2.ShiftID)
	// require.Equal(t, account1.OpeningBalance, account2.OpeningBalance)
	// require.Equal(t, account1.Balance, account2.Balance)
}

func TestListAccounts(t *testing.T) {
	for range 10 {
		createRandomAccount(t)
	}

	arg := ListCashboxAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListCashboxAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)
	for _, account := range accounts {
		// FIXME
		require.NotEmpty(t, account)
	}
}

func TestAddAccountBalance(t *testing.T) {
	account := createRandomAccount(t)
	shift := createRandomShift(t)

	arg := AddCashboxAccountBalanceParams{
		AccountID: account.ID,
		Balance:   util.RandomAmount(),
		ShiftID:   shift.ID,
	}

	account2, err := testQueries.AddCashboxAccountBalance(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, arg.Balance, account2.Balance)

	getBalanceArg := GetCashboxAccountBalanceParams{
		AccountID: account.ID,
		ShiftID:   shift.ID,
	}
	balance, err := testQueries.GetCashboxAccountBalance(context.Background(), getBalanceArg)
	require.NoError(t, err)
	require.Equal(t, account2.Balance, balance.Balance)
}
