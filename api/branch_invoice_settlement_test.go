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

func settlementBody() map[string]any {
	return map[string]any{
		"settlement_ref":  "set-abc",
		"sale_client_ref": "sale-xyz",
		"grand_total":     5000,
		"payments": []map[string]any{
			{"account_name": "Cash", "amount": 3000},
			{"account_name": "Visa", "amount": 2000},
		},
		"occurred_at": time.Now().UTC(),
	}
}

func TestCreateBranchInvoiceSettlement_OK_ReplaysSplitAndNotifies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchInvoiceSettlementByClientRef(gomock.Any(), gomock.Eq(db.GetBranchInvoiceSettlementByClientRefParams{BranchID: testBranchID, ClientRef: "set-abc"})).
		Return(db.BranchInvoiceSettlement{}, sql.ErrNoRows)
	store.EXPECT().
		ApplyBranchSettlementTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(_ context.Context, arg db.ApplyBranchSettlementTxParams) (db.ApplyBranchSettlementTxResult, error) {
			require.Equal(t, "sale-xyz", arg.SaleClientRef)
			require.Len(t, arg.Payments, 2)
			require.Equal(t, int64(3000), arg.Payments[0].Amount)
			return db.ApplyBranchSettlementTxResult{
				Settlement:     db.BranchInvoiceSettlement{ID: 1, ClientRef: arg.ClientRef, SaleClientRef: arg.SaleClientRef},
				InvoiceUpdated: true,
			}, nil
		})
	store.EXPECT().GetBranchInvoiceByClientRef(gomock.Any(), gomock.Any()).
		Return(db.BranchInvoice{ID: 5, BranchInvoiceCode: "SA-000123"}, nil).AnyTimes()
	store.EXPECT().GetDefaultCurrency(gomock.Any()).Return(db.Currency{Code: "USD"}, nil).AnyTimes()

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/invoice_settlements", settlementBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.JSONEq(t, `true`, string(body["invoice_updated"]))
}

func TestCreateBranchInvoiceSettlement_UnknownSaleStillStored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchInvoiceSettlementByClientRef(gomock.Any(), gomock.Any()).Return(db.BranchInvoiceSettlement{}, sql.ErrNoRows)
	store.EXPECT().ApplyBranchSettlementTx(gomock.Any(), gomock.Any()).
		Return(db.ApplyBranchSettlementTxResult{Settlement: db.BranchInvoiceSettlement{ID: 2}, InvoiceUpdated: false}, nil)
	store.EXPECT().GetBranchInvoiceByClientRef(gomock.Any(), gomock.Any()).Return(db.BranchInvoice{}, sql.ErrNoRows).AnyTimes()
	store.EXPECT().GetDefaultCurrency(gomock.Any()).Return(db.Currency{Code: "USD"}, nil).AnyTimes()

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/invoice_settlements", settlementBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.JSONEq(t, `false`, string(body["invoice_updated"]))
}

func TestCreateBranchInvoiceSettlement_IdempotentReplay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchInvoiceSettlementByClientRef(gomock.Any(), gomock.Any()).
		Return(db.BranchInvoiceSettlement{ID: 3, ClientRef: "set-abc"}, nil)
	store.EXPECT().ApplyBranchSettlementTx(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/invoice_settlements", settlementBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchInvoiceSettlement_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchInvoiceSettlementByClientRef(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().ApplyBranchSettlementTx(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/invoice_settlements", map[string]any{"sale_client_ref": "sale-xyz"}))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
