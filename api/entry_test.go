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

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func randomEntry() db.Entry {
	return db.Entry{
		ID:            util.RandomInt(1, 1000),
		CashboxID:     util.RandomInt(1, 1000),
		InventoryID:   util.RandomInt(1, 1000),
		ReferenceType: db.EntryReferenceTypeSalesInvoice,
		ReferenceID:   util.RandomInt(1, 1000),
		Amount:        util.RandomAmount(),
	}
}

func TestCreateEntryAPI(t *testing.T) {
	entry := randomEntry()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"cashbox_id":     entry.CashboxID,
				"inventory_id":   entry.InventoryID,
				"reference_type": string(entry.ReferenceType),
				"reference_id":   entry.ReferenceID,
				"amount":         entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateEntryItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"cashbox_id":     entry.CashboxID,
				"inventory_id":   entry.InventoryID,
				"reference_type": string(entry.ReferenceType),
				"reference_id":   entry.ReferenceID,
				"amount":         entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateEntryItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: map[string]interface{}{
				"cashbox_id": 0,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateEntryItem(gomock.Any(), gomock.Any()).
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

			request, err := http.NewRequest(http.MethodPost, "/entries", bytes.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetEntryAPI(t *testing.T) {
	entry := randomEntry()

	testCases := []struct {
		name          string
		entryID       int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(1).
					Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name:    "NotFound",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(1).
					Return(db.Entry{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(1).
					Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:    "InvalidID",
			entryID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/entries/%d", tc.entryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListEntriesAPI(t *testing.T) {
	n := 5
	entries := make([]db.Entry, n)
	for i := 0; i < n; i++ {
		entries[i] = randomEntry()
	}
	inventoryID := util.RandomInt(1, 1000)

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: fmt.Sprintf("/entries?inventory_id=%d&page_size=5&page_id=1", inventoryID),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Any()).
					Times(1).
					Return(entries, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntries(t, recorder.Body, entries)
			},
		},
		{
			name:  "InternalError",
			query: fmt.Sprintf("/entries?inventory_id=%d&page_size=5&page_id=1", inventoryID),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidPageSize",
			query: fmt.Sprintf("/entries?inventory_id=%d&page_size=1&page_id=1", inventoryID),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Any()).
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

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requiredBodyMatchEntry(t *testing.T, body *bytes.Buffer, entry db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntry db.Entry
	err = json.Unmarshal(data, &gotEntry)
	require.NoError(t, err)
	require.Equal(t, entry.ID, gotEntry.ID)
	require.Equal(t, entry.CashboxID, gotEntry.CashboxID)
	require.Equal(t, entry.InventoryID, gotEntry.InventoryID)
	require.Equal(t, entry.ReferenceType, gotEntry.ReferenceType)
	require.Equal(t, entry.ReferenceID, gotEntry.ReferenceID)
	require.Equal(t, entry.Amount, gotEntry.Amount)
}

func requireBodyMatchEntries(t *testing.T, body *bytes.Buffer, entries []db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntries []db.Entry
	err = json.Unmarshal(data, &gotEntries)
	require.NoError(t, err)
	require.Equal(t, len(entries), len(gotEntries))
	for i := range entries {
		require.Equal(t, entries[i].ID, gotEntries[i].ID)
		require.Equal(t, entries[i].CashboxID, gotEntries[i].CashboxID)
		require.Equal(t, entries[i].InventoryID, gotEntries[i].InventoryID)
	}
}
