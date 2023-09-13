package api

import (
	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	const basePath = "/accounts"
	router.GET(basePath+"/:id", server.getAccount)
	router.GET(basePath, server.listAccount)
	router.POST(basePath, server.createAccount)
	router.PUT(basePath, server.updateAccount)
	router.DELETE(basePath+"/:id", server.deleteAccount)

	server.router = router
	return server
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
