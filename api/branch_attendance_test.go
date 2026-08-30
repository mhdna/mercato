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

func TestCreateBranchAttendanceEvents_InsertsOnlyNewRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	// Row 2 already exists; rows 1 and 3 are new.
	gomock.InOrder(
		store.EXPECT().GetBranchAttendanceEventByClientRef(gomock.Any(), gomock.Eq(db.GetBranchAttendanceEventByClientRefParams{BranchID: testBranchID, ClientRef: "dev|a"})).Return(db.BranchAttendanceEvent{}, sql.ErrNoRows),
		store.EXPECT().CreateBranchAttendanceEvent(gomock.Any(), gomock.Any()).Return(db.BranchAttendanceEvent{ID: 1}, nil),
		store.EXPECT().GetBranchAttendanceEventByClientRef(gomock.Any(), gomock.Eq(db.GetBranchAttendanceEventByClientRefParams{BranchID: testBranchID, ClientRef: "dev|b"})).Return(db.BranchAttendanceEvent{ID: 99}, nil),
		store.EXPECT().GetBranchAttendanceEventByClientRef(gomock.Any(), gomock.Eq(db.GetBranchAttendanceEventByClientRefParams{BranchID: testBranchID, ClientRef: "dev|c"})).Return(db.BranchAttendanceEvent{}, sql.ErrNoRows),
		store.EXPECT().CreateBranchAttendanceEvent(gomock.Any(), gomock.Any()).Return(db.BranchAttendanceEvent{ID: 3}, nil),
	)

	body := map[string]any{"events": []map[string]any{
		{"ref": "dev|a", "salesperson_name": "Ava", "event_date": "2026-08-20", "event_time": "09:00:00", "type": "IN", "status": "Success"},
		{"ref": "dev|b", "salesperson_name": "Ava", "event_date": "2026-08-20", "event_time": "18:00:00", "type": "OUT", "status": "Success"},
		{"ref": "dev|c", "salesperson_name": "Ben", "event_date": "2026-08-20", "event_time": "09:05:00", "type": "IN", "status": "Success"},
	}}

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/attendance_events", body))

	require.Equal(t, http.StatusOK, recorder.Code)
	var out struct {
		Inserted int `json:"inserted"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &out))
	require.Equal(t, 2, out.Inserted)
}

func TestCreateBranchAttendanceEvents_EmptyBatchRejected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)
	store.EXPECT().CreateBranchAttendanceEvent(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/attendance_events", map[string]any{"events": []any{}}))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func attendanceChangeBody(kind string) map[string]any {
	return map[string]any{
		"ref":              "acr-5-complaint",
		"kind":             kind,
		"salesperson_name": "Ben",
		"attendance_date":  "2026-08-21",
		"requested_time":   "09:00",
		"requested_type":   "IN",
		"action":           "replace",
		"status":           "pending",
		"occurred_at":      time.Now().UTC(),
	}
}

func TestCreateBranchAttendanceChange_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().
		GetBranchAttendanceChangeByClientRef(gomock.Any(), gomock.Eq(db.GetBranchAttendanceChangeByClientRefParams{BranchID: testBranchID, ClientRef: "acr-5-complaint"})).
		Return(db.BranchAttendanceChange{}, sql.ErrNoRows)
	store.EXPECT().CreateBranchAttendanceChange(gomock.Any(), gomock.Any()).Times(1).
		Return(db.BranchAttendanceChange{ID: 1, Kind: "complaint"}, nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/attendance_changes", attendanceChangeBody("complaint")))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchAttendanceChange_IdempotentReplay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchAttendanceChangeByClientRef(gomock.Any(), gomock.Any()).
		Return(db.BranchAttendanceChange{ID: 8, Kind: "approval"}, nil)
	store.EXPECT().CreateBranchAttendanceChange(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/attendance_changes", attendanceChangeBody("approval")))

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCreateBranchAttendanceChange_RejectsUnknownKind(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubBranchAuth(store)

	store.EXPECT().GetBranchAttendanceChangeByClientRef(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CreateBranchAttendanceChange(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	server.router.ServeHTTP(recorder, newBranchRequest(t, http.MethodPost, "/branch/attendance_changes", attendanceChangeBody("bogus")))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
