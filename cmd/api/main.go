package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/campuscare/api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	dbPool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Printf("CRITICAL WARNING: Failed to connect to PostgreSQL: %v. The server will start, but database operations will fail until Postgres is running.", err)
		// We continue anyway so the frontend can at least ping the health endpoints
	} else {
		log.Println("Successfully connected to PostgreSQL.")
	}

	srv := server.New(cfg, dbPool)

	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}