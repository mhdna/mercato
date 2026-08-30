package api

import (
	"bytes"
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
	"github.com/mhdna/kashi/util"
	"github.com/stretchr/testify/require"
)

func randomLoan() db.Loan {
	return db.Loan{
		ID:           util.RandomInt(1, 1000),
		Origin:       "central_loan",
		Description:  util.RandomName(),
		CategoryID:   util.RandomInt(1, 10),
		Amount:       util.RandomAmount(),
		CurrencyCode: util.RandomCurrency(),
	}
}

func TestUpdateLoanAPI(t *testing.T) {
	loan := randomLoan()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"id":            loan.ID,
				"description":   loan.Description,
				"category_id":   loan.CategoryID,
				"amount":        loan.Amount,
				"currency_code": loan.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateLoan(gomock.Any(), gomock.Any()).
					Times(1).
					Return(loan, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "MissingID",
			body: map[string]interface{}{
				"description":   loan.Description,
				"category_id":   loan.CategoryID,
				"amount":        loan.Amount,
				"currency_code": loan.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateLoan(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			// A branch-origin loan id matches no rows in UpdateLoan's
			// WHERE origin = 'central_loan' clause -> 404.
			name: "BranchLoanNotEditable",
			body: map[string]interface{}{
				"id":            loan.ID,
				"description":   loan.Description,
				"category_id":   loan.CategoryID,
				"amount":        loan.Amount,
				"currency_code": loan.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateLoan(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Loan{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPut, "/loans", bytes.NewReader(body))
			require.NoError(t, err)
			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteLoanAPI(t *testing.T) {
	loan := randomLoan()

	testCases := []struct {
		name          string
		loanID        int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			loanID: loan.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteLoan(gomock.Any(), gomock.Eq(loan.ID)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			loanID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteLoan(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			loanID: loan.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteLoan(gomock.Any(), gomock.Any()).Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/loans/%d", tc.loanID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)
			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListLoansFilterAPI(t *testing.T) {
	loans := []db.ListLoansRow{
		{ID: 1, Origin: "central_loan", Amount: 5000, PaidAmount: 2000, Status: "partial"},
		{ID: 2, Origin: "central_loan", Amount: 3000, PaidAmount: 3000, Status: "paid"},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().
		ListLoans(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(_ any, arg db.ListLoansParams) ([]db.ListLoansRow, error) {
			require.True(t, arg.CategoryID.Valid)
			require.Equal(t, int64(7), arg.CategoryID.Int64)
			require.Equal(t, "rent", arg.Search.String)
			require.Equal(t, "partial", arg.Status.String)
			return loans, nil
		})
	store.EXPECT().CountLoans(gomock.Any(), gomock.Any()).Times(1).Return(int64(len(loans)), nil)

	server := newTestServer(t, store)
	recorder := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/loans?page_size=10&category_id=7&search=rent&status=partial", nil)
	require.NoError(t, err)
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
}
