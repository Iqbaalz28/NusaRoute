package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"fmt"
	"github.com/nusaroute/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/tracking-service/internal/model"
	"github.com/nusaroute/services/tracking-service/internal/repository"
	"github.com/nusaroute/services/tracking-service/internal/service"
)

func main() {
	logger.InitLogger("tracking-service")
	logger.Info(context.Background(), " Starting NusaRoute Tracking Service...")
	port := getEnv("PORT", "8008")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	mongoDB, err := database.ConnectMongo(database.MongoConfig{
		URI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName: getEnv("MONGO_DB", "nusaroute_tracking"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("MongoDB failed: %v", err)) }

	redisClient, err := database.ConnectRedis(database.RedisConfig{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
	})
	if err != nil { logger.Info(context.Background(), fmt.Sprintf("Redis unavailable: %v", err)) }

	repo := repository.NewTrackingRepository(mongoDB, redisClient)
	trackingSvc := service.NewTrackingService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Kafka consumers — consume ALL status events and record in immutable ledger
	cg := kafka.NewConsumerGroup()

	// Package picked up
	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackagePickedUp, "tracking-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackagePickedUpEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		return trackingSvc.RecordEvent(ctx, model.TrackingEvent{
			AWB: evt.AWB, OrderID: evt.OrderID, Status: "PICKED_UP",
			Location: "Alamat Pengirim", Detail: "Paket telah dijemput kurir",
			Lat: evt.Lat, Lng: evt.Lng, Source: evt.Source, Timestamp: evt.Timestamp,
		})
	})

	// Hub scans
	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageScannedHub, "tracking-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageScannedAtHubEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		detail := map[string]string{"ARRIVED": "Paket tiba di", "SORTED": "Paket disortir di", "DEPARTED": "Paket berangkat dari"}
		return trackingSvc.RecordEvent(ctx, model.TrackingEvent{
			AWB: evt.AWB, OrderID: evt.OrderID, Status: "IN_TRANSIT",
			Location: evt.HubName, Detail: detail[evt.ScanType] + " " + evt.HubName,
			Source: evt.Source, Timestamp: evt.Timestamp,
		})
	})

	// Courier assigned
	cg.Subscribe(ctx, kafkaBrokers, events.TopicCourierAssigned, "tracking-service", func(ctx context.Context, key, value []byte) error {
		var evt events.CourierAssignedEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		return trackingSvc.RecordEvent(ctx, model.TrackingEvent{
			AWB: evt.AWB, OrderID: evt.OrderID, Status: "COURIER_ASSIGNED",
			Location: "-", Detail: "Kurir " + evt.CourierName + " ditugaskan untuk menjemput paket",
			Source: evt.Source, Timestamp: evt.Timestamp,
		})
	})

	// Package delivered
	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageDelivered, "tracking-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageDeliveredEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		return trackingSvc.RecordEvent(ctx, model.TrackingEvent{
			AWB: evt.AWB, OrderID: evt.OrderID, Status: "DELIVERED",
			Location: "Alamat Penerima", Detail: "Paket telah diterima oleh " + evt.ReceiverName,
			Lat: evt.Lat, Lng: evt.Lng, Source: evt.Source, Timestamp: evt.Timestamp,
		})
	})

	// Delivery failed
	cg.Subscribe(ctx, kafkaBrokers, events.TopicDeliveryFailed, "tracking-service", func(ctx context.Context, key, value []byte) error {
		var evt events.DeliveryFailedEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		return trackingSvc.RecordEvent(ctx, model.TrackingEvent{
			AWB: evt.AWB, OrderID: evt.OrderID, Status: "DELIVERY_FAILED",
			Location: "Alamat Penerima", Detail: "Percobaan pengiriman gagal: " + evt.Reason,
			Source: evt.Source, Timestamp: time.Now(),
		})
	})

	defer cg.CloseAll()

	// HTTP endpoints
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Exact match for /api/v1/tracking/recent (must be registered before generic prefix if not using exact Go 1.22 match rules, but Go 1.22 handles this fine)
	mux.HandleFunc("GET /api/v1/tracking/recent", func(w http.ResponseWriter, r *http.Request) {
		eventsList := []model.TrackingEvent{
			{AWB: "NSR123456789", Status: "DELIVERED", Location: "Jakarta Selatan", Timestamp: time.Now().Add(-10 * time.Minute)},
			{AWB: "NSR987654321", Status: "IN_TRANSIT", Location: "Hub Bandung", Timestamp: time.Now().Add(-30 * time.Minute)},
			{AWB: "NSR555555555", Status: "PICKED_UP", Location: "Surabaya", Timestamp: time.Now().Add(-1 * time.Hour)},
		}
		response.Success(w, "recent tracking events retrieved", eventsList)
	})

	mux.HandleFunc("GET /api/v1/tracking/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		awb := parts[len(parts)-1]
		if awb == "health" { response.Success(w, "tracking-service is healthy", map[string]string{"status": "UP"}); return }
		if awb == "live" { return } // handled by next route
		if awb == "recent" { return } // handled by explicit route

		timeline, err := trackingSvc.GetTimeline(r.Context(), awb)
		if err != nil { response.NotFound(w, err.Error()); return }
		response.Success(w, "tracking timeline", timeline)
	})

	mux.HandleFunc("POST /api/v1/tracking/gps", func(w http.ResponseWriter, r *http.Request) {
		var req model.GPSUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		gps := model.CourierGPS{CourierID: req.CourierID, AWB: req.AWB, Lat: req.Lat, Lng: req.Lng}
		if err := trackingSvc.UpdateGPS(r.Context(), gps); err != nil { response.InternalError(w, err.Error()); return }
		response.Success(w, "GPS updated", nil)
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf("✅ Tracking Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil { logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err)) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
