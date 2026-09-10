package api

import (
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

func TestBranchReconcile_ReturnsOnlyMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	// kashi holds ref-a and ref-c; ref-b is missing.
	store.EXPECT().
		BranchInvoiceRefsPresent(gomock.Any(), gomock.Eq(db.BranchInvoiceRefsPresentParams{
			BranchID: testBranchID,
			Column2:  []string{"ref-a", "ref-b", "ref-c"},
		})).
		Return([]string{"ref-a", "ref-c"}, nil)

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, newBranchRequest(t, http.MethodPost, "/branch/sync/reconcile", map[string]any{
		"entity": "invoices",
		"refs":   []string{"ref-a", "ref-b", "ref-c"},
	}))

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Missing []string `json:"missing"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, []string{"ref-b"}, body.Missing)
}

func TestBranchReconcile_RejectsOversizedBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)
	store.EXPECT().BranchInvoiceRefsPresent(gomock.Any(), gomock.Any()).Times(0)

	refs := make([]string, 501)
	for i := range refs {
		refs[i] = "r"
	}

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, newBranchRequest(t, http.MethodPost, "/branch/sync/reconcile", map[string]any{
		"entity": "invoices", "refs": refs,
	}))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBranchInvoice_HistoricalFlagReachesTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchInvoiceByClientRef(gomock.Any(), gomock.Any()).
		Return(db.BranchInvoice{}, sql.ErrNoRows)
	store.EXPECT().
		CreateBranchInvoiceTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, arg db.CreateBranchInvoiceTxParams) (db.CreateBranchInvoiceTxResult, error) {
			require.True(t, arg.Historical, "Historical must pass through to the tx")
			return db.CreateBranchInvoiceTxResult{Invoice: db.BranchInvoice{ID: 1}}, nil
		})
	store.EXPECT().GetDefaultCurrency(gomock.Any()).Return(db.Currency{Code: "USD"}, nil).AnyTimes()

	body := map[string]any{
		"client_ref":                "hist-1",
		"branch_invoice_code":       "H-1",
		"branch_cashbox_account_id": 1,
		"branch_shift_id":           1,
		"branch_inventory_id":       1,
		"occurred_at":               time.Now().UTC(),
		"historical":                true,
		"items": []map[string]any{
			{"branch_product_id": 1, "quantity": 1, "unit_price": 100, "line_total": 100},
		},
	}
	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, newBranchRequest(t, http.MethodPost, "/branch/sales_invoices", body))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPutBranchClient_PhoneMatchDifferentName_StagesConflictAndLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetClientLink(gomock.Any(), gomock.Any()).
		Return(int64(0), sql.ErrNoRows)
	store.EXPECT().
		GetClientByPhone(gomock.Any(), gomock.Eq("70123456")).
		Return(db.Client{ID: 55, Name: "Ali H.", Phone: "70123456", ClientType: "retail"}, nil)
	// Link is still written so invoices attribute immediately.
	store.EXPECT().
		UpsertClientLink(gomock.Any(), gomock.Eq(db.UpsertClientLinkParams{
			BranchID: testBranchID, BranchClientID: 42, ClientID: 55,
		})).
		Return(nil)
	// Name is NOT overwritten.
	store.EXPECT().UpdateClient(gomock.Any(), gomock.Any()).Times(0)
	// Conflict staged.
	store.EXPECT().
		UpsertBranchSyncConflict(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, arg db.UpsertBranchSyncConflictParams) (db.BranchSyncConflict, error) {
			require.Equal(t, "client", arg.Entity)
			require.Equal(t, "42", arg.Ref)
			require.Equal(t, "phone_name_mismatch", arg.Kind)
			require.Equal(t, int64(55), arg.CentralClientID.Int64)
			require.Contains(t, string(arg.BranchPayload), "Ali Hassan")
			return db.BranchSyncConflict{ID: 1}, nil
		})

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, newBranchRequest(t, http.MethodPost, "/branch/clients", map[string]any{
		"branch_client_id": 42, "name": "Ali Hassan", "phone": "70123456",
	}))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"conflict":true`)
}

func TestPutBranchClient_BlankPhone_CreatesWithSyntheticKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	// A blank phone no longer stages a conflict -- it gets a stable
	// synthetic key so the client still syncs and its invoices attribute.
	store.EXPECT().GetClientLink(gomock.Any(), gomock.Any()).Return(int64(0), sql.ErrNoRows)
	store.EXPECT().
		GetClientByPhone(gomock.Any(), gomock.Eq("nophone-7-9")).
		Return(db.Client{}, sql.ErrNoRows)
	store.EXPECT().
		CreateClient(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, arg db.CreateClientParams) (db.Client, error) {
			require.Equal(t, "nophone-7-9", arg.Phone)
			return db.Client{ID: 3, Name: arg.Name, Phone: arg.Phone}, nil
		})
	store.EXPECT().
		UpsertClientLink(gomock.Any(), gomock.Eq(db.UpsertClientLinkParams{
			BranchID: testBranchID, BranchClientID: 9, ClientID: 3,
		})).
		Return(nil)
	store.EXPECT().UpsertBranchSyncConflict(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, newBranchRequest(t, http.MethodPost, "/branch/clients", map[string]any{
		"branch_client_id": 9, "name": "No Phone", "phone": "  ",
	}))

	require.Equal(t, http.StatusOK, rec.Code)
}
