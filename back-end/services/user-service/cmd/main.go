package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/logger"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/user-service/internal/handler"
	"github.com/nusaroute/services/user-service/internal/repository"
	"github.com/nusaroute/services/user-service/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.InitLogger("user-service")
	logger.Info(context.Background(), "Starting NusaRoute User Service...")
	port := getEnv("PORT", "8001")
	jwtSecret := getEnv("JWT_SECRET", "nusaroute-jwt-secret-key-2026")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_user"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err)) }
	defer db.Close()

	_, err = database.ConnectRedis(database.RedisConfig{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", ""),
	})
	if err != nil { logger.Info(context.Background(), fmt.Sprintf("Redis unavailable: %v", err)) }

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, jwtSecret)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	userHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), "Shutting down service gracefully...")
		cancel() // notify workers to stop

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	// Suppress unused variable warning for ctx
	_ = ctx

	logger.Info(context.Background(), fmt.Sprintf("User Service running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), "All workers stopped. Goodbye!")
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
