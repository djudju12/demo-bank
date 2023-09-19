package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/aulas/demo-bank/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTransfer(t *testing.T) {
	amount := int64(10)

	user1, _ := randomUser(t)
	user2, _ := randomUser(t)

	acc1 := randomAccount(user1.Username)
	acc2 := randomAccount(user2.Username)
	acc2.Currency = acc1.Currency

	testCases := []struct {
		baseTestCase //
		request      transferRequest
		setupAuth    func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	}{
		{
			request: transferRequest{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
				Currency:      acc1.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuth(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					expectedArg := db.TransferTxParams{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        amount,
					}

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
						Times(1).
						Return(acc1, nil)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
						Times(1).
						Return(acc2, nil)

					store.EXPECT().
						TransferTx(gomock.Any(), gomock.Eq(expectedArg)).
						Times(1)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			test := newTest(t, "/transfers")
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
