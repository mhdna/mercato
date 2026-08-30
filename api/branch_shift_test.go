package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/stretchr/testify/require"
)

func shiftCloseBody() map[string]any {
	return map[string]any{
		"shift_ref":           "shift-42",
		"branch_shift_id":     42,
		"closing_person_name": "Dana",
		"expected_usd":        10000,
		"counted_usd":         9700,
		"expected_lbp":        4500000,
		"counted_lbp":         4500000,
		"occurred_at":         time.Now().UTC(),
	}
}

func TestCreateBranchShiftClose_OK_DerivesVarianceAndNotifiesOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchShiftByClientRef(gomock.Any(), gomock.Eq(db.GetBranchShiftByClientRefParams{BranchID: testBranchID, ClientRef: "shift-42"})).
		Return(db.BranchShift{}, sql.ErrNoRows)
	store.EXPECT().
		CreateBranchShift(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(_ context.Context, arg db.CreateBranchShiftParams) (db.BranchShift, error) {
			require.Equal(t, int64(-300), arg.VarianceUsd, "variance must be counted - expected")
			require.Equal(t, int64(0), arg.VarianceLbp)
			require.Equal(t, testBranchID, arg.BranchID)
			return db.BranchShift{ID: 1, BranchID: arg.BranchID, ClientRef: arg.ClientRef, VarianceUsd: arg.VarianceUsd}, nil
		})
	store.EXPECT().GetDefaultCurrency(gomock.Any()).Return(db.Currency{Code: "USD"}, nil).AnyTimes()

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/shift_closes", shiftCloseBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchShiftClose_IdempotentReplay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchShiftByClientRef(gomock.Any(), gomock.Any()).
		Return(db.BranchShift{ID: 9, BranchID: testBranchID, ClientRef: "shift-42"}, nil)
	store.EXPECT().CreateBranchShift(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().GetDefaultCurrency(gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/shift_closes", shiftCloseBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchShiftClose_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchShiftByClientRef(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CreateBranchShift(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/shift_closes", map[string]any{"branch_shift_id": 42}))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateBranchShiftClose_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreateBranchShift(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	req := newBranchRequest(t, http.MethodPost, "/branch/shift_closes", shiftCloseBody())
	req.Header.Del(branchKeyHeaderKey)
	server.router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestListBranchShifts_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().ListBranchShifts(gomock.Any(), gomock.Any()).Return([]db.BranchShift{{ID: 1}}, nil)
	store.EXPECT().CountBranchShifts(gomock.Any(), gomock.Any()).Return(int64(1), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/branch_shifts", nil)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Contains(t, body, "branch_shifts")
}
