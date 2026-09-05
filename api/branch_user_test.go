package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/stretchr/testify/require"
)

const buBranchID = int64(7)

func adminRequest(t *testing.T, server *Server, method, url string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	addAuthorization(t, req, server.tokenMaker, authorizationTypeBearer, "operator", time.Minute)

	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, req)
	return rec
}

// stubs the token username -> user id lookup that every command-enqueuing
// handler does for the branch_commands.issued_by FK.
func stubIssuedBy(store *mockdb.MockStore) {
	store.EXPECT().
		GetUserByUsername(gomock.Any(), gomock.Eq("operator")).
		Return(db.User{ID: 99, Name: "operator"}, nil).
		AnyTimes()
}

func TestListBranchUsers_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().
		ListBranchUsersForBranch(gomock.Any(), gomock.Eq(buBranchID)).
		Return([]db.BranchUser{{ID: 1, BranchID: buBranchID, Username: "sara", Role: "cashier", IsActive: true}}, nil)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodGet, fmt.Sprintf("/branches/%d/branch_users", buBranchID), nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var got struct {
		BranchUsers []db.BranchUser `json:"branch_users"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Len(t, got.BranchUsers, 1)
	require.Equal(t, "sara", got.BranchUsers[0].Username)
}

func TestCreateBranchUser_OK_EnqueuesPinCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubIssuedBy(store)

	store.EXPECT().GetBranch(gomock.Any(), gomock.Eq(buBranchID)).Return(db.Branch{ID: buBranchID}, nil)
	store.EXPECT().
		CreateBranchUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, arg db.CreateBranchUserParams) (db.BranchUser, error) {
			require.Equal(t, buBranchID, arg.BranchID)
			require.Equal(t, "sara", arg.Username)
			require.Equal(t, "cashier", arg.Role)
			require.True(t, arg.IsActive)
			return db.BranchUser{ID: 5, BranchID: arg.BranchID, Username: arg.Username, Role: arg.Role, IsActive: true}, nil
		})

	var captured db.CreateBranchCommandParams
	store.EXPECT().
		CreateBranchCommand(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, arg db.CreateBranchCommandParams) (db.BranchCommand, error) {
			captured = arg
			return db.BranchCommand{ID: 1, BranchID: arg.BranchID, Type: arg.Type}, nil
		})

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPost, fmt.Sprintf("/branches/%d/branch_users", buBranchID), map[string]any{
		"username": "sara", "role": "cashier", "pin": "4821",
	})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "set_branch_user_pin", captured.Type)
	require.Equal(t, buBranchID, captured.BranchID)
	require.Equal(t, int64(99), captured.IssuedBy)

	var payload map[string]string
	require.NoError(t, json.Unmarshal(captured.Payload, &payload))
	require.Equal(t, "sara", payload["username"])
	require.Equal(t, "4821", payload["pin"])
}

func TestCreateBranchUser_RejectsShortPin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().CreateBranchUser(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CreateBranchCommand(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPost, fmt.Sprintf("/branches/%d/branch_users", buBranchID), map[string]any{
		"username": "sara", "role": "cashier", "pin": "12",
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateBranchUser_RejectsUnknownRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().CreateBranchUser(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPost, fmt.Sprintf("/branches/%d/branch_users", buBranchID), map[string]any{
		"username": "sara", "role": "superuser", "pin": "4821",
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateBranchUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().GetBranchUser(gomock.Any(), gomock.Eq(int64(5))).
		Return(db.BranchUser{ID: 5, BranchID: buBranchID, Username: "sara"}, nil)
	store.EXPECT().
		UpdateBranchUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, arg db.UpdateBranchUserParams) (db.BranchUser, error) {
			require.Equal(t, int64(5), arg.ID)
			require.Equal(t, "admin", arg.Role)
			require.False(t, arg.IsActive)
			return db.BranchUser{ID: 5, BranchID: buBranchID, Username: arg.Username, Role: arg.Role}, nil
		})

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPut, "/branch_users/5", map[string]any{
		"username": "sara", "role": "admin", "is_active": false,
	})

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSetBranchUserPin_EnqueuesCommandOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	stubIssuedBy(store)

	store.EXPECT().GetBranchUser(gomock.Any(), gomock.Eq(int64(5))).
		Return(db.BranchUser{ID: 5, BranchID: buBranchID, Username: "sara"}, nil)
	store.EXPECT().UpdateBranchUser(gomock.Any(), gomock.Any()).Times(0)

	var captured db.CreateBranchCommandParams
	store.EXPECT().
		CreateBranchCommand(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, arg db.CreateBranchCommandParams) (db.BranchCommand, error) {
			captured = arg
			return db.BranchCommand{ID: 2, Type: arg.Type}, nil
		})

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPut, "/branch_users/5/pin", map[string]any{"pin": "9090"})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "set_branch_user_pin", captured.Type)
	var payload map[string]string
	require.NoError(t, json.Unmarshal(captured.Payload, &payload))
	require.Equal(t, "sara", payload["username"])
	require.Equal(t, "9090", payload["pin"])
}

func TestDeleteBranchUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().GetBranchUser(gomock.Any(), gomock.Eq(int64(5))).
		Return(db.BranchUser{ID: 5, BranchID: buBranchID, Username: "sara"}, nil)
	store.EXPECT().DeleteBranchUser(gomock.Any(), gomock.Eq(int64(5))).Return(nil)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodDelete, "/branch_users/5", nil)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteBranchUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().GetBranchUser(gomock.Any(), gomock.Eq(int64(5))).Return(db.BranchUser{}, sql.ErrNoRows)
	store.EXPECT().DeleteBranchUser(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodDelete, "/branch_users/5", nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPutBranchSettingsManagedLocally_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().
		SetBranchSettingsManagedLocally(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, arg db.SetBranchSettingsManagedLocallyParams) (db.BranchSetting, error) {
			require.Equal(t, buBranchID, arg.BranchID)
			require.ElementsMatch(t, []string{"secrets", "device_ids"}, arg.ManagedLocally)
			return db.BranchSetting{BranchID: arg.BranchID, ManagedLocally: arg.ManagedLocally}, nil
		})

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPut, fmt.Sprintf("/branches/%d/settings/managed_locally", buBranchID), map[string]any{
		"managed_locally": []string{"secrets", "device_ids"},
	})

	require.Equal(t, http.StatusOK, rec.Code)
}
