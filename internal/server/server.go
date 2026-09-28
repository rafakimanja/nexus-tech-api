package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/internal/database"
	"api/internal/domain/uauth"
	"api/internal/domain/user"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port            int
	userService     user.UserService
	userHandler     *user.UserHandler
	userAuthService uauth.UAuthService
	db              database.Service
	logger          *log.Logger
}

func NewServer(
	db database.Service,
	logger *log.Logger,
	userService user.UserService,
	userHandler *user.UserHandler,
	userAuthService uauth.UAuthService,
) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:            port,
		db:              db,
		logger:          logger,
		userService:     userService,
		userHandler:     userHandler,
		userAuthService: userAuthService,
	}

	NewServer.db.DB().AutoMigrate(&user.User{})
	NewServer.db.DB().AutoMigrate(&uauth.UserAuth{})

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
