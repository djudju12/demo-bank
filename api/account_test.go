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

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/aulas/demo-bank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccount(t *testing.T) {
	acc := randomAccount()
	testCase := []struct {
		baseTestCase //
		accountID    int64
	}{
		{
			accountID: acc.ID,
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
						Times(1).
						Return(acc, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, acc)
				},
			},
		},
		{
			accountID: 0,
			baseTestCase: baseTestCase{
				name: "BadRequest",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
		},
		{
			accountID: acc.ID,
			baseTestCase: baseTestCase{
				name: "NotFound",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
						Times(1).
						Return(db.Account{}, sql.ErrNoRows)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusNotFound, recorder.Code)
				},
			},
		},
		{
			accountID: acc.ID,
			baseTestCase: baseTestCase{
				name: "InternalServerError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			// given
			test := newTest(t, fmt.Sprintf("/accounts/%d", tc.accountID))
			tc.buildStubs(test.store)

			request, err := http.NewRequest(http.MethodGet, test.url, nil)
			require.NoError(t, err)

			// when
			test.server.router.ServeHTTP(test.recorder, request)

			// then
			tc.checkResponse(t, test.recorder)
		})

	}

}

func TestCreateAccount(t *testing.T) {
	acc := randomAccount()
	acc.Balance = 0

	testCases := []struct {
		baseTestCase //
		request      createAccountRequest
	}{
		{
			request: createAccountRequest{
				Owner:    acc.Owner,
				Currency: acc.Currency,
			},
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					expectedArg := db.CreateAccountParams{
						Owner:    acc.Owner,
						Currency: acc.Currency,
						Balance:  0,
					}
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Eq(expectedArg)).
						Times(1).
						Return(acc, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, acc)
				},
			},
		},
		{
			request: createAccountRequest{
				Currency: acc.Currency,
			},
			baseTestCase: baseTestCase{
				name: "BadRequest",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
		},
		{
			request: createAccountRequest{
				Owner:    acc.Owner,
				Currency: acc.Currency,
			},
			baseTestCase: baseTestCase{
				name: "InternalServerError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			test := newTest(t, "/accounts")
			tc.buildStubs(test.store)

			body, err := toReader(tc.request)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, test.url, body)
			require.NoError(t, err)

			// when
			test.server.router.ServeHTTP(test.recorder, request)

			// then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	acc := randomAccount()
	testCases := []struct {
		baseTestCase //
		accountID    int64
	}{
		{
			accountID: acc.ID,
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Eq(acc.ID)).
						Times(1).
						Return(nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
				},
			},
		},
		{
			accountID: 0,
			baseTestCase: baseTestCase{
				name: "BadRequest",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
		},
		{
			accountID: acc.ID,
			baseTestCase: baseTestCase{
				name: "NotFound",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(sql.ErrNoRows)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusNotFound, recorder.Code)
				},
			},
		},
		{
			accountID: acc.ID,
			baseTestCase: baseTestCase{
				name: "InternalServerError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			test := newTest(t, fmt.Sprintf("/accounts/%d", tc.accountID))
			tc.buildStubs(test.store)

			request, err := http.NewRequest(http.MethodDelete, test.url, nil)
			require.NoError(t, err)

			// when
			test.server.router.ServeHTTP(test.recorder, request)

			//then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	acc := randomAccount()
	testCases := []struct {
		baseTestCase //
		request      updateAccountRequest
	}{
		{
			request: updateAccountRequest{
				ID:      acc.ID,
				Balance: acc.Balance,
			},
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					expectedArg := db.UpdateAccountParams{
						ID:      acc.ID,
						Balance: acc.Balance,
					}

					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Eq(expectedArg)).
						Times(1).
						Return(acc, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, acc)
				},
			},
		},
		{
			request: updateAccountRequest{
				Balance: acc.Balance,
			},
			baseTestCase: baseTestCase{
				name: "BadRequest",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
		},
		{
			request: updateAccountRequest{
				ID:      acc.ID,
				Balance: acc.Balance,
			},
			baseTestCase: baseTestCase{
				name: "NotFound",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, sql.ErrNoRows)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusNotFound, recorder.Code)
				},
			},
		},
		{
			request: updateAccountRequest{
				ID:      acc.ID,
				Balance: acc.Balance,
			},
			baseTestCase: baseTestCase{
				name: "InternalServerError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			test := newTest(t, "/accounts")
			tc.buildStubs(test.store)

			body, err := toReader(tc.request)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPut, test.url, body)
			require.NoError(t, err)

			// when
			test.server.router.ServeHTTP(test.recorder, request)

			//then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func TestListAccount(t *testing.T) {
	var pageSize int32 = 5
	var pageID int32 = 1
	accs := randomAccountList(pageSize)

	testCases := []struct {
		baseTestCase //
		request      listAccountRequest
	}{
		{
			request: listAccountRequest{
				PageID:   pageID,
				PageSize: pageSize,
			},
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					expectedArg := db.ListAccountParams{
						Limit:  pageSize,
						Offset: (pageID - 1) * pageSize,
					}

					store.EXPECT().
						ListAccount(gomock.Any(), gomock.Eq(expectedArg)).
						Times(1).
						Return(accs, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccountList(t, recorder.Body, accs)
				},
			},
		},
		{
			request: listAccountRequest{
				PageID:   0,
				PageSize: pageSize,
			},
			baseTestCase: baseTestCase{
				name: "BadRequest",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						ListAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
		},
		{
			request: listAccountRequest{
				PageID:   pageID,
				PageSize: pageSize,
			},
			baseTestCase: baseTestCase{
				name: "InternalServerError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						ListAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return([]db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d",
				tc.request.PageID,
				tc.request.PageSize)

			test := newTest(t, url)
			tc.buildStubs(test.store)

			request, err := http.NewRequest(http.MethodGet, test.url, nil)
			require.NoError(t, err)

			// when
			test.server.router.ServeHTTP(test.recorder, request)

			//then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOnwer(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func randomAccountList(pageSize int32) []db.Account {
	var result []db.Account
	for i := 0; i < int(pageSize); i++ {
		result = append(result, randomAccount())
	}

	return result
}

func requireBodyMatchAccountList(t *testing.T, body *bytes.Buffer, accs []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accs, gotAccounts)
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, acc db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, acc, gotAccount)
}

func toReader(body any) (io.Reader, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}
