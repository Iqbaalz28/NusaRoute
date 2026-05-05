package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/payment-service/internal/handler"
	"github.com/nusaroute/services/payment-service/internal/repository"
	"github.com/nusaroute/services/payment-service/internal/service"
)

func main() {
	log.Println("🚀 Starting NusaRoute Payment Service...")

	port := getEnv("PORT", "8002")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "nusaroute"),
		Password: getEnv("DB_PASSWORD", "nusaroute_secret"),
		DBName:   getEnv("DB_NAME", "nusaroute_payment"),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	paymentRepo := repository.NewPaymentRepository(db)
	paymentSvc := service.NewPaymentService(paymentRepo, producer)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	mux := http.NewServeMux()
	paymentHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ Payment Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
