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

func randomPriceList() db.PriceList {
	return db.PriceList{
		ID:        1,
		Name:      "Retail",
		IsActive:  true,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().Add(24 * time.Hour),
	}
}

func TestUpdatePriceListAPI(t *testing.T) {
	pl := randomPriceList()

	testCases := []struct {
		name          string
		body          map[string]any
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]any{
				"id": pl.ID, "name": "Wholesale", "is_active": true, "is_default": false,
				"valid_from": pl.ValidFrom.Format(time.RFC3339),
				"valid_to":   pl.ValidTo.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdatePriceList(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				store.EXPECT().UnsetDefaultPriceList(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "DefaultDemotesOthers",
			body: map[string]any{
				"id": pl.ID, "name": "Wholesale", "is_active": true, "is_default": true,
				"valid_from": pl.ValidFrom.Format(time.RFC3339),
				"valid_to":   pl.ValidTo.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdatePriceList(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				store.EXPECT().UnsetDefaultPriceList(gomock.Any(), gomock.Eq(pl.ID)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "MissingID",
			body: map[string]any{"name": "Wholesale"},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdatePriceList(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
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
			req, err := http.NewRequest(http.MethodPut, "/price_lists", bytes.NewReader(body))
			require.NoError(t, err)
			addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func TestDeletePriceListAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().DeletePriceList(gomock.Any(), gomock.Eq(int64(7))).Times(1).Return(nil)

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodDelete, "/price_lists/7", nil)
	require.NoError(t, err)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSetPriceListBranchesAPI(t *testing.T) {
	testCases := []struct {
		name          string
		listID        string
		body          map[string]any
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			listID: "1",
			body:   map[string]any{"branch_ids": []int64{10, 20}},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetPriceListBranch(gomock.Any(), gomock.Eq(int64(10))).
					Return(db.PriceListBranch{}, sql.ErrNoRows)
				store.EXPECT().GetPriceListBranch(gomock.Any(), gomock.Eq(int64(20))).
					Return(db.PriceListBranch{BranchID: 20, PriceListID: 1}, nil)
				store.EXPECT().DeletePriceListBranchesForList(gomock.Any(), gomock.Eq(int64(1))).Return(nil)
				store.EXPECT().UpsertPriceListBranch(gomock.Any(), db.UpsertPriceListBranchParams{BranchID: 10, PriceListID: 1}).Return(nil)
				store.EXPECT().UpsertPriceListBranch(gomock.Any(), db.UpsertPriceListBranchParams{BranchID: 20, PriceListID: 1}).Return(nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name:   "ConflictWithAnotherList",
			listID: "1",
			body:   map[string]any{"branch_ids": []int64{10}},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetPriceListBranch(gomock.Any(), gomock.Eq(int64(10))).
					Return(db.PriceListBranch{BranchID: 10, PriceListID: 2}, nil)
				store.EXPECT().DeletePriceListBranchesForList(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().UpsertPriceListBranch(gomock.Any(), gomock.Any()).Times(0)
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
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/price_lists/%s/branches", tc.listID), bytes.NewReader(body))
			require.NoError(t, err)
			addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func TestCreatePriceListItemAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreatePriceListItem(gomock.Any(), db.CreatePriceListItemParams{
		PriceListID: 1, ProductID: 5, Price: 2500,
	}).Times(1).Return(db.PriceListItem{PriceListID: 1, ProductID: 5, Price: 2500}, nil)

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	body, err := json.Marshal(map[string]any{"price_list_id": 1, "product_id": 5, "price": 2500})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/price_lists/items", bytes.NewReader(body))
	require.NoError(t, err)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
