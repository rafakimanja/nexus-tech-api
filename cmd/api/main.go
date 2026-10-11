package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"api/internal/auth"
	"api/internal/database"
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/demand"
	"api/internal/domain/teacher"
	"api/internal/domain/uauth"
	"api/internal/domain/user"
	"api/internal/server"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	db := database.New()
	logger := log.Default()

	userRepository := user.NewRepository(db.DB())
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService, logger)

	userAuthRepository := uauth.NewRepository(db.DB())
	userAuthService := uauth.NewService(userAuthRepository)

	demandRepository := demand.NewRepository(db.DB())
	demandService := demand.NewService(demandRepository)
	demandHandler := demand.NewHandler(demandService, logger)

	applicantRepository := applicant.NewRepository(db.DB())
	applicantService := applicant.NewService(applicantRepository)
	applicantHandler := applicant.NewHandler(applicantService, logger)

	areasRepository := areas.NewRepository(db.DB())
	areasService := areas.NewService(areasRepository)
	areasHandler := areas.NewHandler(areasService, logger)

	teacherRepository := teacher.NewRepository(db.DB())
	teacherService := teacher.NewService(teacherRepository, userRepository)
	teacherHandler := teacher.NewHandler(teacherService, logger)

	if err := auth.NewAuth(); err != nil {
		log.Fatalf("authentication configuration error: %v", err)
	}

	server := server.NewServer(
		db,
		logger,
		userService,
		userHandler,
		userAuthService,
		demandService,
		demandHandler,
		applicantService,
		applicantHandler,
		areasService,
		areasHandler,
		teacherService,
		teacherHandler,
	)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
