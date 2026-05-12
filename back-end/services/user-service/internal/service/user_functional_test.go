//go:build functional

package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/services/user-service/internal/model"
	"github.com/nusaroute/services/user-service/internal/repository"
	"github.com/nusaroute/services/user-service/internal/service"
)

func TestUserFunctional(t *testing.T) {
	// Setup real database connection
	// These env vars should be provided by the CI/CD environment or local docker-compose
	host := os.Getenv("DB_HOST")
	if host == "" { host = "localhost" }
	port := os.Getenv("DB_PORT")
	if port == "" { port = "5432" }
	user := os.Getenv("DB_USER")
	if user == "" { user = "postgres" }
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" { pass = "postgres" }
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "nusaroute_user" }

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: host, Port: port, User: user, Password: pass, DBName: dbname,
	})
	if err != nil {
		t.Skipf("Skipping functional test: failed to connect to DB: %v", err)
		return
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)

	ctx := context.Background()
	testEmail := "functional@nusaroute.id"

	// Cleanup if exists
	_, _ = db.Exec("DELETE FROM users WHERE email = $1", testEmail)

	t.Run("Register", func(t *testing.T) {
		user, err := svc.Register(ctx, model.RegisterRequest{
			Email: testEmail, Password: "SecurePass123!", FullName: "Functional Test User",
		})
		if err != nil { t.Fatalf("failed to register: %v", err) }
		if user.Email != testEmail { t.Errorf("expected %s, got %s", testEmail, user.Email) }
	})

	t.Run("Login", func(t *testing.T) {
		res, err := svc.Login(ctx, model.LoginRequest{
			Email: testEmail, Password: "SecurePass123!",
		}, "test-secret")
		if err != nil { t.Fatalf("failed to login: %v", err) }
		if res.Token == "" { t.Error("expected token") }
	})

	t.Run("GetProfile", func(t *testing.T) {
		// We need to get the user ID first
		u, _ := repo.GetByEmail(ctx, testEmail)
		profile, err := svc.GetProfile(ctx, u.ID)
		if err != nil { t.Fatalf("failed to get profile: %v", err) }
		if profile.Email != testEmail { t.Errorf("expected %s, got %s", testEmail, profile.Email) }
	})

	// Final cleanup
	_, _ = db.Exec("DELETE FROM users WHERE email = $1", testEmail)
}
