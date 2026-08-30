package db

import (
	"context"
	"testing"
	"time"

	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func createRandomCoupon(t *testing.T) Coupon {
	client := createRandomClient(t)
	arg := CreateCouponParams{
		Code:         util.RandomCode(),
		ClientID:     client.ID,
		Status:       CouponStatusActive,
		DiscountType: DiscountTypeFixed,
		Reason:       util.RandomString(50),
		ValidUntil:   time.Now().Add(time.Minute * time.Duration(util.RandomNumber())),
	}

	coupon, err := testQueries.CreateCoupon(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, coupon)
	require.Equal(t, arg.Code, coupon.Code)
	require.Equal(t, arg.ClientID, coupon.ClientID)
	require.Equal(t, arg.Status, coupon.Status)
	require.Equal(t, arg.Reason, coupon.Reason)
	require.WithinDuration(t, arg.ValidUntil, coupon.ValidUntil, time.Second)

	return coupon
}

func TestCreateCoupon(t *testing.T) {
	createRandomCurrency(t)
}

func TestGetCoupon(t *testing.T) {
	coupon1 := createRandomCoupon(t)
	coupon2, err := testQueries.GetCoupon(context.Background(), coupon1.Code)

	require.NoError(t, err)
	require.NotEmpty(t, coupon2)

	require.Equal(t, coupon1.Code, coupon2.Code)
	require.Equal(t, coupon1.Code, coupon2.Code)
	require.Equal(t, coupon1.ClientID, coupon2.ClientID)
	require.Equal(t, coupon1.Status, coupon2.Status)
	require.Equal(t, coupon1.Reason, coupon2.Reason)
	require.WithinDuration(t, coupon1.ValidUntil, coupon2.ValidUntil, time.Second)
}

func TestDeactivateCoupon(t *testing.T) {
	coupon1 := createRandomCoupon(t)
	err := testQueries.DeactivateCoupon(context.Background(), coupon1.Code)
	require.NoError(t, err)

	coupon2, err := testQueries.GetCoupon(context.Background(), coupon1.Code)
	require.Equal(t, coupon1.Code, coupon2.Code)
	require.Equal(t, coupon2.Status, CouponStatusInactive)
}

func TestListCoupons(t *testing.T) {
	for range 10 {
		createRandomCoupon(t)
	}
	arg := ListCouponsParams{
		PageSize:   5,
		PageOffset: 5,
	}
	coupons, err := testQueries.ListCoupons(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, coupons, 5)
	for _, coupon := range coupons {
		require.NotEmpty(t, coupon)
	}
}
