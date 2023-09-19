package api

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	mockdb "github.com/aulas/demo-bank/db/mock"
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/aulas/demo-bank/util"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

type test struct {
	ctrl     *gomock.Controller
	store    *mockdb.MockStore
	server   *Server
	recorder *httptest.ResponseRecorder
	url      string
}

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		TokenDuration:     time.Minute,
	}

	server, err := NewServer(&config, store)
	require.NoError(t, err)

	return server
}

func newTest(t *testing.T, url string) *test {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)

	return &test{
		ctrl:     ctrl,
		store:    store,
		server:   server,
		recorder: httptest.NewRecorder(),
		url:      url,
	}
}

type baseTestCase struct {
	name          string
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
