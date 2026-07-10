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

func TestGetInventoryAPI(t *testing.T) {
	inventory := randomInventory()

	testCases := []struct {
		name          string
		inventoryID   int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			inventoryID: inventory.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetInventory(gomock.Any(), gomock.Eq(inventory.ID)).
					Times(1).
					Return(inventory, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchInventory(t, recorder.Body, inventory)
			},
		},
		{
			name:        "NotFound",
			inventoryID: inventory.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetInventory(gomock.Any(), gomock.Eq(inventory.ID)).
					Times(1).
					Return(db.Inventory{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:        "InternalError",
			inventoryID: inventory.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetInventory(gomock.Any(), gomock.Eq(inventory.ID)).
					Times(1).
					Return(db.Inventory{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "InvalidID",
			inventoryID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetInventory(gomock.Any(), gomock.Any()).
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

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start the test server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/inventories/%d", tc.inventoryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomInventory() db.Inventory {
	return db.Inventory{
		ID:        util.RandomInt(1, 1000),
		Type:      db.InventoryTypeStore,
		Code:      util.RandomCode(),
		Latitude:  util.RandomLongitudeLatitude(),
		Longitude: util.RandomLongitudeLatitude(),
	}
}

func requiredBodyMatchInventory(t *testing.T, body *bytes.Buffer, inventory db.Inventory) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotInventory db.Inventory
	err = json.Unmarshal(data, &gotInventory)
	require.NoError(t, err)
	require.Equal(t, inventory, gotInventory)
}
