package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

const (
	testBranchCode = "BR01"
	testBranchKey  = "branch-secret-key"
	testBranchID   = int64(7)
)

// stubBranchAuth wires the mock so branchAuthMiddleware resolves
// testBranchCode/testBranchKey to an active branch with id testBranchID.
func stubBranchAuth(store *mockdb.MockStore) {
	hash, _ := util.HashPassword(testBranchKey)
	store.EXPECT().
		GetBranchByCode(gomock.Any(), gomock.Eq(testBranchCode)).
		Return(db.Branch{ID: testBranchID, Code: testBranchCode, ApiKeyHash: hash, IsActive: true}, nil).
		AnyTimes()
	store.EXPECT().
		UpdateBranchLastSeenAt(gomock.Any(), gomock.Eq(testBranchID)).
		Return(nil).
		AnyTimes()
}

func newBranchRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, path, reader)
	require.NoError(t, err)
	req.Header.Set(branchCodeHeaderKey, testBranchCode)
	req.Header.Set(branchKeyHeaderKey, testBranchKey)
	req.Header.Set("Content-Type", "application/json")
	return req
}
