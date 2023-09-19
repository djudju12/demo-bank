package api

import (
	"fmt"

	db "github.com/aulas/demo-bank/db/sqlc"
	"github.com/aulas/demo-bank/token"
	"github.com/aulas/demo-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	config     *util.Config
}

func NewServer(config *util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	router := gin.Default()
	authRouter := router.Group("/").Use(authMiddleware(server.tokenMaker))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	const accountsPath = "/accounts"
	authRouter.GET(path(accountsPath, "/:id"), server.getAccount)
	authRouter.GET(accountsPath, server.listAccount)
	authRouter.POST(accountsPath, server.createAccount)
	authRouter.PUT(accountsPath, server.updateAccount)
	authRouter.DELETE(path(accountsPath, "/:id"), server.deleteAccount)

	const transfersPath = "/transfers"
	authRouter.POST(transfersPath, server.createTransfer)

	const usersPath = "/users"
	router.POST(usersPath, server.createUser)
	router.POST(path(usersPath, "/login"), server.loginUser)

	server.router = router
	return server, nil
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
