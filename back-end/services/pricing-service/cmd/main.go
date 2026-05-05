package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/pricing-service/internal/handler"
	"github.com/nusaroute/services/pricing-service/internal/repository"
	"github.com/nusaroute/services/pricing-service/internal/service"
)

func main() {
	log.Println("🚀 Starting NusaRoute Pricing Service...")
	port := getEnv("PORT", "8003")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", "nusaroute_secret"),
		DBName: getEnv("DB_NAME", "nusaroute_pricing"),
	})
	if err != nil { log.Fatalf("DB failed: %v", err) }
	defer db.Close()

	redisClient, err := database.ConnectRedis(database.RedisConfig{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", "nusaroute_secret"),
	})
	if err != nil { log.Printf("Redis unavailable (caching disabled): %v", err) }

	pricingRepo := repository.NewPricingRepository(db)
	pricingSvc := service.NewPricingService(pricingRepo, redisClient)
	pricingHandler := handler.NewPricingHandler(pricingSvc)

	mux := http.NewServeMux()
	pricingHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ Pricing Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil { log.Fatalf("Server failed: %v", err) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
