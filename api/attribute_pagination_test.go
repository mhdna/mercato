package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestListAttributeValuesPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	values := []db.ListAttributeValuesPageRow{{
		ID: 8, AttributeID: 2, AttributeName: "brand", Value: "Acme",
	}}

	store.EXPECT().ListAttributeValuesPage(gomock.Any(), db.ListAttributeValuesPageParams{
		AttributeName: "brand",
		Search:        "ac",
		SortBy:        "created_at",
		SortOrder:     "desc",
		PageOffset:    14,
		PageSize:      14,
	}).Return(values, nil)
	store.EXPECT().CountAttributeValues(gomock.Any(), db.CountAttributeValuesParams{
		AttributeName: "brand",
		Search:        "ac",
	}).Return(int64(23), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(
		http.MethodGet,
		"/attributes/?attribute=brand&page_size=14&page_id=14&search=ac&sort_by=created_at&sort_order=desc",
		nil,
	)
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"values":[{"id":8,"attribute_id":2,"value":"Acme","created_at":"0001-01-01T00:00:00Z","attribute_name":"brand"}],"total":23}`, recorder.Body.String())
}

func TestListAttributeValuesRequiresAttribute(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/attributes/?page_size=14", nil)
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
