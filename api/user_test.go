package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/aulas/demo-bank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.ComparePasswords(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matchs arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)
	testCases := []struct {
		baseTestCase //
		request      createUserRequest
	}{
		{
			request: createUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			baseTestCase: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					}

					store.EXPECT().
						CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
						Times(1).
						Return(user, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusCreated, recorder.Code)
					requireBodyMatchUser(t, recorder.Body, user)
				},
			}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			test := newTest(t, "/users")
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

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(12)

	return db.User{
		Username:       util.RandomOnwer(),
		HashedPassword: password,
		FullName:       util.RandomFullName(),
		Email:          util.RandomEmail(),
	}, password
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)

	require.Empty(t, gotUser.HashedPassword)

	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.WithinDuration(t, user.PasswordChangedAt, gotUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user.CreatedAt, gotUser.CreatedAt, time.Second)
}
