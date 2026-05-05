package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/courier-service/internal/handler"
	"github.com/nusaroute/services/courier-service/internal/repository"
	"github.com/nusaroute/services/courier-service/internal/service"
)

func main() {
	log.Println("🚀 Starting NusaRoute Courier Service...")
	port := getEnv("PORT", "8005")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", "nusaroute_secret"),
		DBName: getEnv("DB_NAME", "nusaroute_courier"),
	})
	if err != nil { log.Fatalf("DB failed: %v", err) }
	defer db.Close()

	courierRepo := repository.NewCourierRepository(db)
	courierSvc := service.NewCourierService(courierRepo)
	courierHandler := handler.NewCourierHandler(courierSvc)

	mux := http.NewServeMux()
	courierHandler.RegisterRoutes(mux)

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ Courier Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil { log.Fatalf("Server failed: %v", err) }
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
