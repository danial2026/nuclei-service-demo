package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/repository/postgres"
	"nuclei-service-demo/internal/server"
	"nuclei-service-demo/internal/service"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load env file
	if err := godotenv.Load(); err != nil {
		log.Printf("[%s] Warning: .env file not found, using environment variables", time.Now().Format(time.RFC3339))
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database connection
	db, err := postgres.NewConnection(cfg.DB)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	scanRepo := postgres.NewScanRepository(db, cfg, logger)

	// Initialize services
	nucleiService := service.NewNucleiService(cfg, logger)

	// Initialize and start scan worker
	scanWorker := service.NewScanWorker(scanRepo, nucleiService, logger)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go scanWorker.Start(workerCtx)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("[%s] Warning: .env file not found, using environment variables", time.Now().Format(time.RFC3339))
	}

	// Create server
	log.Printf("[%s] Initializing server...", time.Now().Format(time.RFC3339))
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("[%s] Failed to create server: %v", time.Now().Format(time.RFC3339), err)
	}
	log.Printf("[%s] Server initialized successfully", time.Now().Format(time.RFC3339))

	go func() {
		// Start server
		log.Printf("[%s] Starting server on %s:%d...", time.Now().Format(time.RFC3339), cfg.Server.Host, cfg.Server.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("[%s] Failed to start server: %v", time.Now().Format(time.RFC3339), err)
		}
	}()

	go func() {
		// Start demo server
		log.Printf("[%s] Initializing Demo server...", time.Now().Format(time.RFC3339))
		demoSrv, err := server.NewDemoServer(cfg)
		if err != nil {
			log.Fatalf("[%s] Failed to create demo server: %v", time.Now().Format(time.RFC3339), err)
		}
		if err := demoSrv.Start(); err != nil {
			log.Fatalf("[%s] Failed to start demo server: %v", time.Now().Format(time.RFC3339), err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	logger.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
}
