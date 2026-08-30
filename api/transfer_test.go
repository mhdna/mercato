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

func randomTransfer() db.Transfer {
	return db.Transfer{
		ID:              util.RandomInt(1, 1000),
		FromInventoryID: util.RandomInt(1, 1000),
		ToInventoryID:   util.RandomInt(1, 1000),
		Type:            db.TransferTypeAssets,
	}
}

func TestCreateTransferAPI(t *testing.T) {
	transfer := randomTransfer()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"from_inventory_id": transfer.FromInventoryID,
				"to_inventory_id":   transfer.ToInventoryID,
				"type":              string(transfer.Type),
				"items": []map[string]interface{}{
					{"variant_id": 1, "quantity": 3},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateTransferTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateTransferTxResult{Transfer: transfer}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"from_inventory_id": transfer.FromInventoryID,
				"to_inventory_id":   transfer.ToInventoryID,
				"type":              string(transfer.Type),
				"items": []map[string]interface{}{
					{"variant_id": 1, "quantity": 3},
				},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateTransferTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateTransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: map[string]interface{}{
				"from_inventory_id": 0,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateTransferTx(gomock.Any(), gomock.Any()).
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

			request, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(body))
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetTransferAPI(t *testing.T) {
	transfer := randomTransfer()

	testCases := []struct {
		name          string
		transferID    int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(transfer, nil)
				store.EXPECT().
					ListTransferItems(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return([]db.ListTransferItemsRow{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},
		{
			name:       "NotFound",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfer{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			transferID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/transfers/%d", tc.transferID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListTransfersAPI(t *testing.T) {
	n := 5
	transfers := make([]db.ListTransfersRow, n)
	for i := 0; i < n; i++ {
		tr := randomTransfer()
		transfers[i] = db.ListTransfersRow{
			ID:              tr.ID,
			FromInventoryID: tr.FromInventoryID,
			ToInventoryID:   tr.ToInventoryID,
			Type:            tr.Type,
		}
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "/transfers?page_size=5&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Any()).
					Times(1).
					Return(transfers, nil)
				store.EXPECT().
					CountTransfers(gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfers(t, recorder.Body, transfers)
			},
		},
		{
			name:  "InternalError",
			query: "/transfers?page_size=5&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidPageSize",
			query: "/transfers?page_size=1&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Any()).
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

func TestUpdateTransferAPI(t *testing.T) {
	transfer := randomTransfer()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"id":                transfer.ID,
				"from_inventory_id": transfer.FromInventoryID,
				"to_inventory_id":   transfer.ToInventoryID,
				"type":              string(transfer.Type),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateTransfer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"id":                transfer.ID,
				"from_inventory_id": transfer.FromInventoryID,
				"to_inventory_id":   transfer.ToInventoryID,
				"type":              string(transfer.Type),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateTransfer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: map[string]interface{}{
				"id": 0,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateTransfer(gomock.Any(), gomock.Any()).
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

			request, err := http.NewRequest(http.MethodPut, "/transfers", bytes.NewReader(body))
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requiredBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	// createTransfer/getTransfer respond with {"transfer": {...}, "items": [...]}.
	var resp struct {
		Transfer db.Transfer `json:"transfer"`
	}
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	gotTransfer := resp.Transfer
	require.Equal(t, transfer.ID, gotTransfer.ID)
	require.Equal(t, transfer.FromInventoryID, gotTransfer.FromInventoryID)
	require.Equal(t, transfer.ToInventoryID, gotTransfer.ToInventoryID)
	require.Equal(t, transfer.Type, gotTransfer.Type)
}

func requireBodyMatchTransfers(t *testing.T, body *bytes.Buffer, transfers []db.ListTransfersRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	// listTransfers responds with {"transfers": [...], "total": ...}, not a bare array.
	var resp struct {
		Transfers []db.ListTransfersRow `json:"transfers"`
	}
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	gotTransfers := resp.Transfers
	require.Equal(t, len(transfers), len(gotTransfers))
	for i := range transfers {
		require.Equal(t, transfers[i].ID, gotTransfers[i].ID)
		require.Equal(t, transfers[i].FromInventoryID, gotTransfers[i].FromInventoryID)
		require.Equal(t, transfers[i].ToInventoryID, gotTransfers[i].ToInventoryID)
	}
}
