package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise-erp/core/internal/config"
	"github.com/enterprise-erp/core/internal/database"
	"github.com/enterprise-erp/core/internal/queue"
	"github.com/enterprise-erp/core/internal/server"
)

func main() {
	cfg := config.Load()
	log.Printf("[INFO] Initializing Enterprise Headless ERP Backend on port :%s (env: %s)...", cfg.Port, cfg.Environment)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Database Pool
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Failed to connect to database: %v. Running in disconnected mode.", err)
	} else {
		defer db.Close()
	}

	// Initialize PostgreSQL SKIP LOCKED Background Worker Pool
	var workerPool *queue.WorkerPool
	if db != nil && db.Pool != nil {
		workerPool = queue.NewWorkerPool(db, 4)
		workerPool.Start()
		defer workerPool.Stop()
	}

	// Create Chi HTTP Router with IDP & SSE endpoints
	router := server.NewRouter(cfg, db, workerPool)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server runner in background goroutine
	go func() {
		log.Printf("[INFO] Rocket-Speed ERP HTTP Server listening at http://localhost:%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	log.Println("[INFO] Shutting down ERP server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[ERROR] Server forced to shutdown: %v", err)
	}

	log.Println("[INFO] ERP server exited safely.")
}
