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

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/courier-service/internal/handler"
	"github.com/nusaroute/services/courier-service/internal/repository"
	"github.com/nusaroute/services/courier-service/internal/service"
)

func main() {
	logger.InitLogger("courier-service")
	logger.Info(context.Background(), fmt.Sprint("🚀 Starting NusaRoute Courier Service..."))
	port := getEnv("PORT", "8005")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_courier"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err)) }
	defer db.Close()

	courierRepo := repository.NewCourierRepository(db)
	courierSvc := service.NewCourierService(courierRepo)
	courierHandler := handler.NewCourierHandler(courierSvc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	courierHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf("✅ Courier Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil { logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err)) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
