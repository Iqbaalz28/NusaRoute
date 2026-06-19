package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/logger"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/analytics-service/internal/repository"
)

func main() {
	logger.InitLogger("analytics-service")
	logger.Info(context.Background(), " Starting NusaRoute Analytics Service (Data Warehouse)...")
	port := getEnv("PORT", "8011")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_warehouse"),
	})
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err))
	}
	defer db.Close()

	repo := repository.NewAnalyticsRepository(db)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Each endpoint maps to an OLAP aggregate over the star schema.
	mux.HandleFunc("GET /api/v1/analytics/overview", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.Overview(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "overview retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/monthly", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.Monthly(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "monthly trend retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/by-province", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.ByProvince(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "by-province retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/by-service", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.ByService(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "by-service retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/by-status", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.ByStatus(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "by-status retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/driver", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.Driver(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "driver metrics retrieved", data)
	})
	// ---- ML analytics (predictive layer over the same star schema) ----
	mux.HandleFunc("GET /api/v1/analytics/ml/forecast", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.Forecast(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "forecast retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/ml/segments", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.Segments(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "segments retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/ml/risk", func(w http.ResponseWriter, r *http.Request) {
		data, err := repo.Risk(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "risk retrieved", data)
	})
	mux.HandleFunc("GET /api/v1/analytics/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "analytics-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf("✅ Analytics Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
