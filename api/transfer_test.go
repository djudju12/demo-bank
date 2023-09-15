package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTransfer(t *testing.T) {
	amount := int64(10)

	acc1 := randomAccount()
	acc2 := randomAccount()
	acc2.Currency = acc1.Currency

	testCases := []struct {
		baseTestCase //
		request      transferRequest
	}{
		{
			request: transferRequest{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
				Currency:      acc1.Currency,
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
			test.server.router.ServeHTTP(test.recorder, request)

			// then
			tc.checkResponse(t, test.recorder)
		})
	}
}
