package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	mockdb "github.com/mhdna/kashi/db/mock"
	"github.com/stretchr/testify/require"
)

func TestBulkDeleteColorsAPI(t *testing.T) {
	testCases := []struct {
		name       string
		body       map[string]any
		buildStubs func(store *mockdb.MockStore)
		wantCode   int
	}{
		{
			name: "OK",
			body: map[string]any{"ids": []int64{1, 2, 3}},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteColors(gomock.Any(), gomock.Eq([]int64{1, 2, 3})).Times(1).Return(int64(3), nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:       "EmptyIDs",
			body:       map[string]any{"ids": []int64{}},
			buildStubs: func(store *mockdb.MockStore) { store.EXPECT().DeleteColors(gomock.Any(), gomock.Any()).Times(0) },
			wantCode:   http.StatusBadRequest,
		},
		{
			name: "ForeignKeyConflict",
			body: map[string]any{"ids": []int64{5}},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteColors(gomock.Any(), gomock.Any()).Times(1).
					Return(int64(0), &pq.Error{Code: "23503"})
			},
			wantCode: http.StatusConflict,
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

			request, err := http.NewRequest(http.MethodPost, "/colors/bulk_delete", bytes.NewReader(body))
			require.NoError(t, err)
			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			require.Equal(t, tc.wantCode, recorder.Code)
		})
	}
}

func TestBulkDeleteCurrenciesAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().DeleteCurrencies(gomock.Any(), gomock.Eq([]string{"AAA", "BBB"})).Times(1).Return(int64(2), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]any{"codes": []string{"AAA", "BBB"}})
	request, err := http.NewRequest(http.MethodPost, "/currencies/bulk_delete", bytes.NewReader(body))
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestBulkDeletePriceListItemsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().
		DeletePriceListItems(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(_ any, arg any) (int64, error) {
			return 2, nil
		})

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]any{"product_ids": []int64{10, 11}})
	request, err := http.NewRequest(http.MethodPost, "/price_lists/7/items/bulk_delete", bytes.NewReader(body))
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
}
