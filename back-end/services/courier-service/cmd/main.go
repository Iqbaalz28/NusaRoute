package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/nusaroute/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	grpc_server "github.com/nusaroute/services/courier-service/internal/grpc"
	"github.com/nusaroute/services/courier-service/internal/handler"
	"github.com/nusaroute/services/courier-service/internal/repository"
	"github.com/nusaroute/services/courier-service/internal/service"
)

func main() {
	logger.InitLogger("courier-service")
	logger.Info(context.Background(), fmt.Sprint(" Starting NusaRoute Courier Service..."))
	port := getEnv("PORT", "8005")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_courier"),
	})
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err))
	}
	defer db.Close()

	courierRepo := repository.NewCourierRepository(db)
	courierSvc := service.NewCourierService(courierRepo)
	courierHandler := handler.NewCourierHandler(courierSvc)

	var wg sync.WaitGroup
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	courierHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h) // surface X-User-ID (from gateway) into context
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf(" Courier Service running on port %s", port))

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(":"+port, h); err != nil {
			logger.Log.Fatal(fmt.Sprintf("HTTP Server failed: %v", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcPort := getEnv("GRPC_PORT", "9005")
		logger.Info(context.Background(), fmt.Sprintf(" gRPC Server running on port %s", grpcPort))
		if err := grpc_server.StartGRPCServer(":" + grpcPort); err != nil {
			logger.Log.Fatal(fmt.Sprintf("gRPC Server failed: %v", err))
		}
	}()

	wg.Wait()
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
