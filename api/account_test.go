package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/aulas/demo-bank/token"
	"github.com/aulas/demo-bank/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccount(t *testing.T) {
	user, _ := randomUser(t)
	acc := randomAccount(user.Username)

	testCase := []struct {
		baseTestCase //
		accountID    int64
		setupAuth    func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			accountID: acc.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			tc.setupAuth(t, request, test.server.tokenMaker)
			test.server.router.ServeHTTP(test.recorder, request)

			// then
			tc.checkResponse(t, test.recorder)
		})

	}

}

func TestCreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	acc := randomAccount(user.Username)
	acc.Balance = 0

	testCases := []struct {
		baseTestCase //
		request      createAccountRequest
		setupAuth    func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			request: createAccountRequest{
				Currency: acc.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
				Currency: "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
				Currency: acc.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			baseTestCase: baseTestCase{
				name: "foreign_key_violation",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, &pq.Error{Code: "23503"})
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusForbidden, recorder.Code)
				},
			},
		},
		{
			request: createAccountRequest{
				Currency: acc.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			baseTestCase: baseTestCase{
				name: "unique_violation",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, &pq.Error{Code: "23505"})
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusForbidden, recorder.Code)
				},
			},
		},
		{
			request: createAccountRequest{
				Currency: acc.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			baseTestCase: baseTestCase{
				name: "StatusInternalServerError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, errors.New("some error"))
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
		},
		{
			request: createAccountRequest{
				Currency: acc.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// ...
			},
			baseTestCase: baseTestCase{
				name: "Unauthorized",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
			tc.setupAuth(t, request, test.server.tokenMaker)
			test.server.router.ServeHTTP(test.recorder, request)

			// then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	user, _ := randomUser(t)
	acc := randomAccount(user.Username)
	testCases := []struct {
		baseTestCase //
		accountID    int64
		setupAuth    func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			accountID: acc.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
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
			tc.setupAuth(t, request, test.server.tokenMaker)
			test.server.router.ServeHTTP(test.recorder, request)

			//then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	user, _ := randomUser(t)
	acc := randomAccount(user.Username)

	testCases := []struct {
		baseTestCase //
		request      updateAccountRequest
		setupAuth    func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			request: updateAccountRequest{
				ID:      acc.ID,
				Balance: acc.Balance,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
			tc.setupAuth(t, request, test.server.tokenMaker)
			test.server.router.ServeHTTP(test.recorder, request)

			//then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func TestListAccount(t *testing.T) {
	var pageSize int32 = 5
	var pageID int32 = 1
	user, _ := randomUser(t)

	accs := make([]db.Account, pageSize)
	for i := 0; i < int(pageSize); i++ {
		accs[i] = randomAccount(user.Username)
	}

	testCases := []struct {
		baseTestCase //
		request      listAccountRequest
		setupAuth    func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			request: listAccountRequest{
				PageID:   pageID,
				PageSize: pageSize,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					expectedArg := db.ListAccountParams{
						Owner:  accs[0].Owner,
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
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
			tc.setupAuth(t, request, test.server.tokenMaker)
			test.server.router.ServeHTTP(test.recorder, request)

			//then
			tc.checkResponse(t, test.recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
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
