package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/internal/database"
	"api/internal/domain/user"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(),
	}
	
	NewServer.db.DB().AutoMigrate(&user.User{})
	NewServer.db.DB().AutoMigrate(&user.UserAuth{})

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
