package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"context"
	"fmt"
	"github.com/nusaroute/pkg/logger"
	"net/http"
	"os"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/pricing-service/internal/handler"
	"github.com/nusaroute/services/pricing-service/internal/repository"
	"github.com/nusaroute/services/pricing-service/internal/service"
)

func main() {
	logger.InitLogger("pricing-service")
	logger.Info(context.Background(), fmt.Sprint(" Starting NusaRoute Pricing Service..."))
	port := getEnv("PORT", "8003")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_pricing"),
	})
	if err != nil { logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err)) }
	defer db.Close()

	redisClient, err := database.ConnectRedis(database.RedisConfig{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", ""),
	})
	if err != nil { logger.Info(context.Background(), fmt.Sprintf("Redis unavailable (caching disabled): %v", err)) }

	pricingRepo := repository.NewPricingRepository(db)
	pricingSvc := service.NewPricingService(pricingRepo, redisClient)
	pricingHandler := handler.NewPricingHandler(pricingSvc)


	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	pricingHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf("✅ Pricing Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil { logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err)) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
