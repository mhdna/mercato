package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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

func randomExpense() db.Expense {
	return db.Expense{
		ID:           util.RandomInt(1, 1000),
		Description:  util.RandomName(),
		Amount:       util.RandomAmount(),
		CurrencyCode: util.RandomCurrency(),
	}
}

func TestCreateExpenseAPI(t *testing.T) {
	expense := randomExpense()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"description":   expense.Description,
				"amount":        expense.Amount,
				"currency_code": expense.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateExpense(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expense, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchExpense(t, recorder.Body, expense)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"description":   expense.Description,
				"amount":        expense.Amount,
				"currency_code": expense.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateExpense(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Expense{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			body: map[string]interface{}{
				"description": "",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateExpense(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			request, err := http.NewRequest(http.MethodPost, "/expenses", bytes.NewReader(body))
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetExpenseAPI(t *testing.T) {
	expense := randomExpense()

	testCases := []struct {
		name          string
		expenseID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			expenseID: expense.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetExpense(gomock.Any(), gomock.Eq(expense.ID)).
					Times(1).
					Return(expense, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchExpense(t, recorder.Body, expense)
			},
		},
		{
			name:      "NotFound",
			expenseID: expense.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetExpense(gomock.Any(), gomock.Eq(expense.ID)).
					Times(1).
					Return(db.Expense{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			expenseID: expense.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetExpense(gomock.Any(), gomock.Eq(expense.ID)).
					Times(1).
					Return(db.Expense{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			expenseID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetExpense(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			url := fmt.Sprintf("/expenses/%d", tc.expenseID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListExpensesAPI(t *testing.T) {
	n := 5
	expenses := make([]db.Expense, n)
	for i := 0; i < n; i++ {
		expenses[i] = randomExpense()
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "/expenses?page_size=5&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expenses, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchExpenses(t, recorder.Body, expenses)
			},
		},
		{
			name:  "InternalError",
			query: "/expenses?page_size=5&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidPageSize",
			query: "/expenses?page_size=1&page_id=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListExpenses(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			request, err := http.NewRequest(http.MethodGet, tc.query, nil)
			require.NoError(t, err)

			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requiredBodyMatchExpense(t *testing.T, body *bytes.Buffer, expense db.Expense) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotExpense db.Expense
	err = json.Unmarshal(data, &gotExpense)
	require.NoError(t, err)
	require.Equal(t, expense.ID, gotExpense.ID)
	require.Equal(t, expense.Description, gotExpense.Description)
	require.Equal(t, expense.Amount, gotExpense.Amount)
	require.Equal(t, expense.CurrencyCode, gotExpense.CurrencyCode)
}

func requireBodyMatchExpenses(t *testing.T, body *bytes.Buffer, expenses []db.Expense) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotExpenses []db.Expense
	err = json.Unmarshal(data, &gotExpenses)
	require.NoError(t, err)
	require.Equal(t, len(expenses), len(gotExpenses))
	for i := range expenses {
		require.Equal(t, expenses[i].ID, gotExpenses[i].ID)
		require.Equal(t, expenses[i].Description, gotExpenses[i].Description)
		require.Equal(t, expenses[i].Amount, gotExpenses[i].Amount)
	}
}

func TestUpdateExpenseAPI(t *testing.T) {
	expense := randomExpense()

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"id":            expense.ID,
				"description":   expense.Description,
				"amount":        expense.Amount,
				"currency_code": expense.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateExpense(gomock.Any(), gomock.Any()).
					Times(1).
					Return(expense, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchExpense(t, recorder.Body, expense)
			},
		},
		{
			name: "MissingID",
			body: map[string]interface{}{
				"description":   expense.Description,
				"amount":        expense.Amount,
				"currency_code": expense.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateExpense(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: map[string]interface{}{
				"id":            expense.ID,
				"description":   expense.Description,
				"amount":        expense.Amount,
				"currency_code": expense.CurrencyCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateExpense(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Expense{}, sql.ErrNoRows)
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

			request, err := http.NewRequest(http.MethodPut, "/expenses", bytes.NewReader(body))
			require.NoError(t, err)
			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteExpenseAPI(t *testing.T) {
	expense := randomExpense()

	testCases := []struct {
		name          string
		expenseID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			expenseID: expense.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteExpense(gomock.Any(), gomock.Eq(expense.ID)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			expenseID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteExpense(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			expenseID: expense.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteExpense(gomock.Any(), gomock.Any()).Times(1).Return(sql.ErrConnDone)
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

			url := fmt.Sprintf("/expenses/%d", tc.expenseID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)
			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, "user", time.Minute)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
