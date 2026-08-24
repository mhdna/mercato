package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func randomPurchase() db.Purchase {
	return db.Purchase{
		ID:          util.RandomInt(1, 1000),
		SupplierID:  util.RandomInt(1, 1000),
		PurchasedAt: time.Now(),
	}
}

func TestCreatePurchaseAPI(t *testing.T) {
	purchase := randomPurchase()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"supplier_id":  purchase.SupplierID,
				"purchased_at": purchase.PurchasedAt.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePurchase(gomock.Any(), gomock.Any()).
					Times(1).
					Return(purchase, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchPurchase(t, recorder.Body, purchase)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"supplier_id":  purchase.SupplierID,
				"purchased_at": purchase.PurchasedAt.Format(time.RFC3339),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePurchase(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Purchase{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: map[string]interface{}{
				"supplier_id": 0,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePurchase(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
			recorder := httptest.NewRecorder()

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/purchases", bytes.NewReader(body))
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetPurchaseAPI(t *testing.T) {
	purchase := randomPurchase()

	testCases := []struct {
		name          string
		purchaseID    int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			purchaseID: purchase.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPurchase(gomock.Any(), gomock.Eq(purchase.ID)).
					Times(1).
					Return(purchase, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchPurchase(t, recorder.Body, purchase)
			},
		},
		{
			name:       "NotFound",
			purchaseID: purchase.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPurchase(gomock.Any(), gomock.Eq(purchase.ID)).
					Times(1).
					Return(db.Purchase{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			purchaseID: purchase.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPurchase(gomock.Any(), gomock.Eq(purchase.ID)).
					Times(1).
					Return(db.Purchase{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			purchaseID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPurchase(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/purchases/%d", tc.purchaseID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListPurchasesAPI(t *testing.T) {
	n := 5
	purchases := make([]db.Purchase, n)
	for i := 0; i < n; i++ {
		purchases[i] = randomPurchase()
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "/purchases?page_size=5&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPurchases(gomock.Any(), gomock.Any()).
					Times(1).
					Return(purchases, nil)
				store.EXPECT().
					CountPurchases(gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPurchases(t, recorder.Body, purchases)
			},
		},
		{
			name:  "InternalError",
			query: "/purchases?page_size=5&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPurchases(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidPageSize",
			query: "/purchases?page_size=1&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPurchases(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, tc.query, nil)
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requiredBodyMatchPurchase(t *testing.T, body *bytes.Buffer, purchase db.Purchase) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPurchase db.Purchase
	err = json.Unmarshal(data, &gotPurchase)
	require.NoError(t, err)
	require.Equal(t, purchase.ID, gotPurchase.ID)
	require.Equal(t, purchase.SupplierID, gotPurchase.SupplierID)
}

func requireBodyMatchPurchases(t *testing.T, body *bytes.Buffer, purchases []db.Purchase) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	// listPurchases responds with {"purchases": [...], "total": ...}, not a bare array.
	var resp struct {
		Purchases []db.Purchase `json:"purchases"`
	}
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	gotPurchases := resp.Purchases
	require.Equal(t, len(purchases), len(gotPurchases))
	for i := range purchases {
		require.Equal(t, purchases[i].ID, gotPurchases[i].ID)
		require.Equal(t, purchases[i].SupplierID, gotPurchases[i].SupplierID)
	}
}
