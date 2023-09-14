package api

import (
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	const accountsPath = "/accounts"
	router.GET(path(accountsPath, "/:id"), server.getAccount)
	router.GET(path(accountsPath), server.listAccount)
	router.POST(path(accountsPath), server.createAccount)
	router.PUT(path(accountsPath), server.updateAccount)
	router.DELETE(path(accountsPath, "/:id"), server.deleteAccount)

	const transfersPath = "/transfers"
	router.POST(path(transfersPath), server.createTransfer)

	server.router = router
	return server
}

func path(strings ...string) string {
	var result string
	for _, s := range strings {
		result += s
	}

	return result
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
