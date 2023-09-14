package api

import (
	"net/http/httptest"
	"os"
	"testing"

	mockdb "github.com/aulas/demo-bank/db/mock"

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

func newTest(t *testing.T, url string) *test {
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	return &test{
		ctrl:     ctrl,
		store:    store,
		server:   NewServer(store),
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
