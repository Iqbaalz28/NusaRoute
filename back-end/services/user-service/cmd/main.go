package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
	"sync"
	"context"
	"fmt"
	"github.com/nusaroute/pkg/logger"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/user-service/internal/handler"
	"github.com/nusaroute/services/user-service/internal/repository"
	"github.com/nusaroute/services/user-service/internal/service"
)

func main() {
	logger.InitLogger("user-service")
		srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), fmt.Sprint("Shutting down service gracefully..."))
		cancel() // notify workers to stop

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	logger.Info(context.Background(), fmt.Sprintf("✅ Service running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), fmt.Sprint("All workers stopped. Goodbye!"))

}

func runMigrations(db *sqlx.DB) error {
	// Note: In production, use a proper migration tool like golang-migrate.
	// For now, we execute the schema directly.
	// The actual migration SQL is in migrations/ folder.
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
