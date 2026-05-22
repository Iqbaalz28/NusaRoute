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
	"os/signal"
	"strings"
	"syscall"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/order-service/internal/handler"
	"github.com/nusaroute/services/order-service/internal/repository"
	"github.com/nusaroute/services/order-service/internal/service"
)

func main() {
	logger.InitLogger("order-service")
	logger.Info(context.Background(), fmt.Sprint("🚀 Starting NusaRoute Order Service..."))

	port := getEnv("PORT", "8004")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_order"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("DB connection failed: %v", err)) }
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo, producer)
	orderHandler := handler.NewOrderHandler(orderSvc)

	// Start background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// SLA Monitor: detect stuck packages (48h no update → LOST_SUSPECTED)
	go orderSvc.RunSLAMonitor(ctx)

	// Saga: auto-cancel orders pending payment > 24h
	go orderSvc.RunPaymentExpiryChecker(ctx)

	// Kafka consumers
	consumerGroup := kafka.NewConsumerGroup()

	// Subscribe to payment.success → update order to READY_FOR_PICKUP
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicPaymentSuccess, "order-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PaymentSuccessEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		logger.Info(context.Background(), fmt.Sprintf("[Order] Received payment.success for order=%s", evt.OrderID))
		return orderSvc.HandlePaymentSuccess(ctx, evt.OrderID)
	})

	// Subscribe to delivery.failed → increment attempts / return-to-sender
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicDeliveryFailed, "order-service", func(ctx context.Context, key, value []byte) error {
		var evt events.DeliveryFailedEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		logger.Info(context.Background(), fmt.Sprintf("[Order] Received delivery.failed for order=%s (attempt #%d)", evt.OrderID, evt.AttemptNum))
		return orderSvc.HandleDeliveryFailed(ctx, evt.OrderID, evt.AttemptNum)
	})

	// Subscribe to package.delivered → mark order as delivered
	consumerGroup.Subscribe(ctx, kafkaBrokers, events.TopicPackageDelivered, "order-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageDeliveredEvent
		if err := json.Unmarshal(value, &evt); err != nil { return err }
		logger.Info(context.Background(), fmt.Sprintf("[Order] Received package.delivered for order=%s", evt.OrderID))
		return orderSvc.HandlePackageDelivered(ctx, evt.OrderID)
	})

	defer consumerGroup.CloseAll()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	orderHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
