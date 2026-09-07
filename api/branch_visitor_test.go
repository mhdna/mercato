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

func visitorEventBody() map[string]any {
	return map[string]any{
		"ref":         "vis-uuid-1",
		"day":         "2026-09-07",
		"direction":   "in",
		"occurred_at": time.Now().UTC(),
	}
}

func TestCreateBranchVisitorEvent_OK_InsertsOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchVisitorEventByClientRef(gomock.Any(), gomock.Eq(db.GetBranchVisitorEventByClientRefParams{BranchID: testBranchID, ClientRef: "vis-uuid-1"})).
		Return(db.BranchVisitorEvent{}, sql.ErrNoRows)
	store.EXPECT().
		CreateBranchVisitorEvent(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(_ any, arg db.CreateBranchVisitorEventParams) (db.BranchVisitorEvent, error) {
			require.Equal(t, testBranchID, arg.BranchID)
			require.Equal(t, "in", arg.Direction)
			require.Equal(t, "2026-09-07", arg.Day)
			return db.BranchVisitorEvent{ID: 1, BranchID: arg.BranchID, ClientRef: arg.ClientRef, Direction: arg.Direction}, nil
		})

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/visitor_events", visitorEventBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchVisitorEvent_IdempotentReplay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchVisitorEventByClientRef(gomock.Any(), gomock.Any()).
		Return(db.BranchVisitorEvent{ID: 9, BranchID: testBranchID, ClientRef: "vis-uuid-1"}, nil)
	store.EXPECT().CreateBranchVisitorEvent(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/visitor_events", visitorEventBody()))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchVisitorEvent_BadDirection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchVisitorEventByClientRef(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CreateBranchVisitorEvent(gomock.Any(), gomock.Any()).Times(0)

	body := visitorEventBody()
	body["direction"] = "sideways"

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/visitor_events", body))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateBranchVisitorEvent_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreateBranchVisitorEvent(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	req := newBranchRequest(t, http.MethodPost, "/branch/visitor_events", visitorEventBody())
	req.Header.Del(branchKeyHeaderKey)
	server.router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestListBranchVisitorDays_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().ListBranchVisitorDays(gomock.Any(), gomock.Any()).
		Return([]db.ListBranchVisitorDaysRow{{BranchID: 1, Day: "2026-09-07", InCount: 40, OutCount: 12}}, nil)
	store.EXPECT().CountBranchVisitorDays(gomock.Any(), gomock.Any()).Return(int64(1), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/branch_visitor_events", nil)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)
	server.router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Contains(t, body, "visitor_days")
}
