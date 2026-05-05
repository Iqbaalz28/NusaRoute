package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/hub-service/internal/model"
	"github.com/nusaroute/services/hub-service/internal/repository"
	"github.com/nusaroute/services/hub-service/internal/service"
)

func main() {
	log.Println("🚀 Starting NusaRoute Hub Service...")
	port := getEnv("PORT", "8007")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", "nusaroute_secret"),
		DBName: getEnv("DB_NAME", "nusaroute_hub"),
	})
	if err != nil { log.Fatalf("DB failed: %v", err) }
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	hubRepo := repository.NewHubRepository(db)
	hubSvc := service.NewHubService(hubRepo, producer)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/hub/scan/inbound", func(w http.ResponseWriter, r *http.Request) {
		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		req.ScanType = "ARRIVED"
		scan, err := hubSvc.Scan(r.Context(), req)
		if err != nil { response.BadRequest(w, err.Error()); return }
		response.Created(w, "inbound scan recorded", scan)
	})
	mux.HandleFunc("POST /api/v1/hub/scan/outbound", func(w http.ResponseWriter, r *http.Request) {
		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		req.ScanType = "DEPARTED"
		scan, err := hubSvc.Scan(r.Context(), req)
		if err != nil { response.BadRequest(w, err.Error()); return }
		response.Created(w, "outbound scan recorded", scan)
	})
	mux.HandleFunc("POST /api/v1/hub/scan/sort", func(w http.ResponseWriter, r *http.Request) {
		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		req.ScanType = "SORTED"
		scan, err := hubSvc.Scan(r.Context(), req)
		if err != nil { response.BadRequest(w, err.Error()); return }
		response.Created(w, "sort scan recorded", scan)
	})
	mux.HandleFunc("GET /api/v1/hub/manifest/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		hubID := parts[len(parts)-1]
		manifest, err := hubSvc.GetManifest(r.Context(), hubID)
		if err != nil { response.InternalError(w, err.Error()); return }
		response.Success(w, "manifest retrieved", manifest)
	})
	mux.HandleFunc("GET /api/v1/hub/list", func(w http.ResponseWriter, r *http.Request) {
		hubs, err := hubSvc.ListHubs(r.Context())
		if err != nil { response.InternalError(w, err.Error()); return }
		response.Success(w, "hubs retrieved", hubs)
	})
	mux.HandleFunc("GET /api/v1/hub/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "hub-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ Hub Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil { log.Fatalf("Server failed: %v", err) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
