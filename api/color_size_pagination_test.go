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

func TestListColorsPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	colors := []db.Color{{ID: 9, Name: "Blue", HexValue: "#0000FF"}}

	store.EXPECT().ListColorsPage(gomock.Any(), db.ListColorsPageParams{
		Search: "blue", SortBy: "created_at", SortOrder: "desc",
		PageOffset: 14, PageSize: 14,
	}).Return(colors, nil)
	store.EXPECT().CountColors(gomock.Any(), "blue").Return(int64(27), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/colors?page_size=14&page_id=14&search=blue&sort_by=created_at&sort_order=desc", nil)
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"colors":[{"id":9,"name":"Blue","hex_value":"#0000FF","version":0,"created_at":"0001-01-01T00:00:00Z"}],"total":27}`, recorder.Body.String())
}

func TestListSizesPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	sizes := []db.Size{{ID: 4, Name: "XL", Type: "Shirt", Order: "4"}}

	store.EXPECT().ListSizesPage(gomock.Any(), db.ListSizesPageParams{
		Search: "shirt", SortBy: "created_at", SortOrder: "desc",
		PageOffset: 28, PageSize: 14,
	}).Return(sizes, nil)
	store.EXPECT().CountSizes(gomock.Any(), "shirt").Return(int64(31), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/sizes?page_size=14&page_id=28&search=shirt&sort_by=created_at&sort_order=desc", nil)
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"sizes":[{"id":4,"name":"XL","type":"Shirt","order":"4","version":0,"created_at":"0001-01-01T00:00:00Z"}],"total":31}`, recorder.Body.String())
}
