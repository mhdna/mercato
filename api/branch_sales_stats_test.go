package api

import (
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

func TestListBranchSalesStats_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	from, _ := time.Parse("2006-01-02", "2026-01-01")
	to, _ := time.Parse("2006-01-02", "2026-09-30")
	day, _ := time.Parse("2006-01-02", "2026-09-07")

	store.EXPECT().
		ListBranchSalesDaysRange(gomock.Any(), gomock.Eq(db.ListBranchSalesDaysRangeParams{FromDay: from, ToDay: to})).
		Return([]db.ListBranchSalesDaysRangeRow{{BranchID: 1, Day: day, InvoiceCount: 18, Revenue: 540000}}, nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/branch_sales_stats?from=2026-01-01&to=2026-09-30", nil)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Contains(t, body, "sales_days")

	var days []branchSalesDayEntry
	require.NoError(t, json.Unmarshal(body["sales_days"], &days))
	require.Len(t, days, 1)
	require.Equal(t, "2026-09-07", days[0].Day)
	require.Equal(t, int64(18), days[0].InvoiceCount)
	require.Equal(t, int64(540000), days[0].Revenue)
}

func TestListBranchSalesStats_RejectsBadRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().ListBranchSalesDaysRange(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/branch_sales_stats?from=2026-09-30&to=2026-01-01", nil)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
