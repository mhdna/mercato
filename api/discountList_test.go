package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestDeleteDiscountListAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().DeleteDiscountList(gomock.Any(), gomock.Eq(int64(3))).Times(1).Return(nil)

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodDelete, "/discount_lists/3", nil)
	require.NoError(t, err)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateDiscountListItemZeroDiscountAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreateDiscountListItem(gomock.Any(), db.CreateDiscountListItemParams{
		DiscountListID: 1, ProductID: 5, Discount: 0,
	}).Times(1).Return(db.DiscountListItem{DiscountListID: 1, ProductID: 5, Discount: 0}, nil)

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	body, err := json.Marshal(map[string]any{"discount_list_id": 1, "product_id": 5, "discount": 0})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/discount_lists/items", bytes.NewReader(body))
	require.NoError(t, err)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSetDiscountListBranchesAPI(t *testing.T) {
	testCases := []struct {
		name          string
		body          map[string]any
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]any{"branch_ids": []int64{10}},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetDiscountListBranch(gomock.Any(), gomock.Eq(int64(10))).
					Return(db.DiscountListBranch{}, sql.ErrNoRows)
				store.EXPECT().DeleteDiscountListBranchesForList(gomock.Any(), gomock.Eq(int64(1))).Return(nil)
				store.EXPECT().UpsertDiscountListBranch(gomock.Any(), db.UpsertDiscountListBranchParams{BranchID: 10, DiscountListID: 1}).Return(nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "ConflictWithAnotherList",
			body: map[string]any{"branch_ids": []int64{10}},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetDiscountListBranch(gomock.Any(), gomock.Eq(int64(10))).
					Return(db.DiscountListBranch{BranchID: 10, DiscountListID: 2}, nil)
				store.EXPECT().DeleteDiscountListBranchesForList(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().UpsertDiscountListBranch(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, rec.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			rec := httptest.NewRecorder()

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/discount_lists/%d/branches", 1), bytes.NewReader(body))
			require.NoError(t, err)
			addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}
