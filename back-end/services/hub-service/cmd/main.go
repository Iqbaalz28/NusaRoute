package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nusaroute/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/hub-service/internal/model"
	"github.com/nusaroute/services/hub-service/internal/repository"
	"github.com/nusaroute/services/hub-service/internal/service"
	"github.com/nusaroute/services/hub-service/internal/stream"
)

func main() {
	logger.InitLogger("hub-service")
	logger.Info(context.Background(), " Starting NusaRoute Hub Service...")
	port := getEnv("PORT", "8007")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_hub"),
	})
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err))
	}
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	hubRepo := repository.NewHubRepository(db)
	hubSvc := service.NewHubService(hubRepo, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Start Stateful Stream Processor (Tumbling Window)
	windowSize := 60 * time.Second // 1 minute tumbling window
	analyticsStream := stream.NewHubAnalyticsStream(kafkaBrokers, producer, windowSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		analyticsStream.Start(ctx)
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("POST /api/v1/hub/scan/inbound", func(w http.ResponseWriter, r *http.Request) {
		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		req.ScanType = "ARRIVED"
		scan, err := hubSvc.Scan(r.Context(), req)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		response.Created(w, "inbound scan recorded", scan)
	})
	mux.HandleFunc("POST /api/v1/hub/scan/outbound", func(w http.ResponseWriter, r *http.Request) {
		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		req.ScanType = "DEPARTED"
		scan, err := hubSvc.Scan(r.Context(), req)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		response.Created(w, "outbound scan recorded", scan)
	})
	mux.HandleFunc("POST /api/v1/hub/scan/sort", func(w http.ResponseWriter, r *http.Request) {
		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		req.ScanType = "SORTED"
		scan, err := hubSvc.Scan(r.Context(), req)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		response.Created(w, "sort scan recorded", scan)
	})
	mux.HandleFunc("GET /api/v1/hub/manifest/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		hubID := parts[len(parts)-1]
		manifest, err := hubSvc.GetManifest(r.Context(), hubID)
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "manifest retrieved", manifest)
	})
	mux.HandleFunc("GET /api/v1/hub/list", func(w http.ResponseWriter, r *http.Request) {
		hubs, err := hubSvc.ListHubs(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "hubs retrieved", hubs)
	})
	mux.HandleFunc("GET /api/v1/hub/stats", func(w http.ResponseWriter, r *http.Request) {
		activeHubs, totalCities, err := hubSvc.GetDashboardStats(r.Context())
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "stats retrieved", map[string]interface{}{
			"active_hubs":  activeHubs,
			"total_cities": totalCities,
		})
	})
	mux.HandleFunc("GET /api/v1/hub/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "hub-service is healthy", map[string]string{"status": "UP"})
	})

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

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
		}
	}()

	// Graceful shutdown logic needed for stream processor to stop cleanly
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), "Shutting down Hub Service gracefully...")
		cancel()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	logger.Info(context.Background(), fmt.Sprintf(" Hub Service running on port %s", port))

	wg.Wait()
	logger.Info(context.Background(), "Hub Service stopped cleanly.")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
