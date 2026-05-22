package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
	"sync"
	"fmt"
	"github.com/nusaroute/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/dispatch-service/internal/repository"
	"github.com/nusaroute/services/dispatch-service/internal/service"
)

func main() {
	logger.InitLogger("dispatch-service")
	logger.Info(context.Background(), fmt.Sprint("🚀 Starting NusaRoute Dispatch Service..."))
	port := getEnv("PORT", "8006")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	courierSvcURL := getEnv("COURIER_SERVICE_URL", "http://localhost:8005")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_dispatch"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err)) }
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	dispatchRepo := repository.NewDispatchRepository(db)
	dispatchSvc := service.NewDispatchService(dispatchRepo, producer, courierSvcURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Background: monitor courier no-shows
	go dispatchSvc.RunNoShowMonitor(ctx)

	// Kafka: subscribe to order.ready-for-pickup → auto-assign courier
	consumerGroup := kafka.NewConsumerGroup()
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicOrderReadyForPickup, "dispatch-service",
		func(ctx context.Context, key, value []byte) error {
			var evt events.OrderReadyForPickupEvent
			if err := json.Unmarshal(value, &evt); err != nil { return err }
			logger.Info(context.Background(), fmt.Sprintf("[Dispatch] Order ready for pickup: %s", evt.OrderID))
			return dispatchSvc.AutoAssign(ctx, evt.OrderID, evt.AWB, evt.PickupLat, evt.PickupLng, evt.PickupAddr)
		})
	defer consumerGroup.CloseAll()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("GET /api/v1/dispatch/assignments", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		assignments, err := dispatchSvc.ListAssignments(r.Context(), status, 1, 50)
		if err != nil { response.InternalError(w, err.Error()); return }
		response.Success(w, "assignments retrieved", assignments)
	})
	mux.HandleFunc("GET /api/v1/dispatch/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "dispatch-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf("✅ Dispatch Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil { logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err)) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
