package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomShift(t *testing.T) Shift {
	cashbox := createRandomCashbox(t)

	// arg := CreateShiftParams{
	// 	CashboxID: cashbox.ID,
	// 	// TotalOpeningBalance: util.RandomAmount(),
	// 	// TotalBalance:        util.RandomAmount(),
	// }

	shift, err := testQueries.CreateShift(context.Background(), cashbox.ID)
	require.NoError(t, err)
	require.Equal(t, shift.CashboxID, cashbox.ID)
	// require.Equal(t, shift.CashboxID, arg.CashboxID)
	// require.Equal(t, shift.TotalOpeningBalance, arg.TotalOpeningBalance)
	// require.Equal(t, shift.TotalBalance, arg.TotalBalance)

	return shift
}

func TestCreateShift(t *testing.T) {
	createRandomShift(t)
}

func TestGetShift(t *testing.T) {
	shift1 := createRandomShift(t)

	shift2, err := testQueries.GetShift(context.Background(), shift1.ID)
	require.NoError(t, err)
	require.Equal(t, shift1.ID, shift2.ID)
	require.Equal(t, shift1.CashboxID, shift2.CashboxID)
	// require.Equal(t, shift1.TotalOpeningBalance, shift2.TotalOpeningBalance)
	// require.Equal(t, shift1.TotalBalance, shift2.TotalBalance)

	require.WithinDuration(t, shift1.CreatedAt, shift2.CreatedAt, time.Second)
	require.Equal(t, shift1.ClosedAt.Valid, shift2.ClosedAt.Valid)
}

func TestListShifts(t *testing.T) {
	for range 10 {
		createRandomShift(t)
	}

	arg := ListShiftsParams{
		SortBy:     "id",
		SortOrder:  "desc",
		PageSize:   5,
		PageOffset: 5,
	}

	shifts, err := testQueries.ListShifts(context.Background(), arg)
	require.NoError(t, err)
	for _, shift := range shifts {
		// FIXME
		require.NotEmpty(t, shift)
	}
}

// func TestUpdateShiftBalance(t *testing.T) {
// 	shift1 := createRandomShift(t)

// 	arg := AddToShiftBalanceParams{
// 		ID:     shift1.ID,
// 		Amount: util.RandomAmount(),
// 	}

// 	_, err := testQueries.AddToShiftBalance(context.Background(), arg)
// 	require.NoError(t, err)

// 	shift2, err := testQueries.GetShift(context.Background(), shift1.ID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	require.NoError(t, err)
// 	require.Equal(t, shift2.TotalBalance, arg.Amount+shift1.TotalBalance)
// }

// func TestCloseShift(t *testing.T) {
// 	shift := createRandomShift(t)
//
// 	arg := CloseShiftParams{
// 		ID:              shift.ID,
// 		ClosingDateTime: sql.NullTime{Time: time.Now(), Valid: true},
// 		IsClosed:        true,
// 	}
//
// 	err := testQueries.CloseShift(context.Background(), arg)
// 	require.NoError(t, err)
//
// 	shift2, err := testQueries.GetShift(context.Background(), shift.ID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	require.NoError(t, err)
// 	require.Equal(t, shift2.IsClosed, arg.IsClosed)
// 	require.WithinDuration(t, shift2.ClosingDateTime.Time, arg.ClosingDateTime.Time, time.Second)
// }

func TestCloseShift(t *testing.T) {
	shift := createRandomShift(t)

	err := testQueries.CloseShift(context.Background(), shift.ID)
	require.NoError(t, err)

	closedShift, err := testQueries.GetShift(context.Background(), shift.ID)
	require.NoError(t, err)
	require.True(t, closedShift.IsClosed)
	require.True(t, closedShift.ClosedAt.Valid)
}
