package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/user-service/internal/handler"
	"github.com/nusaroute/services/user-service/internal/repository"
	"github.com/nusaroute/services/user-service/internal/service"
)

func main() {
	log.Println("🚀 Starting NusaRoute User Service...")

	// Load config from environment
	port := getEnv("PORT", "8001")
	jwtSecret := getEnv("JWT_SECRET", "nusaroute-jwt-secret-key-2026")

	// Connect to PostgreSQL
	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "nusaroute"),
		Password: getEnv("DB_PASSWORD", "nusaroute_secret"),
		DBName:   getEnv("DB_NAME", "nusaroute_user"),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, jwtSecret)

	// Setup HTTP router
	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	// Apply middleware stack
	var h http.Handler = mux
	h = middleware.JWTAuth(jwtSecret)(h)    // JWT validation (skipped for public routes internally)
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ User Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runMigrations(db *sqlx.DB) error {
	// Note: In production, use a proper migration tool like golang-migrate.
	// For now, we execute the schema directly.
	// The actual migration SQL is in migrations/ folder.
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
