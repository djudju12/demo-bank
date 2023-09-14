package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/gin-gonic/gin"
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
		request      gin.H
	}{
		{
			request: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          amount,
				"currency":        acc1.Currency,
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
	}
}
