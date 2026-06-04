package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/logger"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/services/payment-service/internal/handler"
	"github.com/nusaroute/services/payment-service/internal/repository"
	"github.com/nusaroute/services/payment-service/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.InitLogger("payment-service")
	logger.Info(context.Background(), "Starting NusaRoute Payment Service...")
	port := getEnv("PORT", "8002")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_payment"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err)) }
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	paymentRepo := repository.NewPaymentRepository(db)
	paymentSvc := service.NewPaymentService(paymentRepo, producer)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Start Outbox Worker — polls outbox_events table and publishes to Kafka
	outboxWorker := outbox.NewWorker(db, producer, 2*time.Second)
	outboxWorker.Start(ctx)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	paymentHandler.RegisterRoutes(mux)

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

	logger.Info(context.Background(), fmt.Sprintf("Payment Service running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), "All workers stopped. Goodbye!")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
