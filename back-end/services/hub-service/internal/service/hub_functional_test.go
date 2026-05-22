//go:build functional

package service_test

import (
	"os"
	"testing"

	"github.com/nusaroute/pkg/database"
)

func TestHubFunctional(t *testing.T) {
	host := os.Getenv("DB_HOST")
	if host == "" { host = "localhost" }
	port := os.Getenv("DB_PORT")
	if port == "" { port = "5432" }
	user := os.Getenv("DB_USER")
	if user == "" { user = "postgres" }
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" { pass = "postgres" }
	dbname := os.Getenv("DB_NAME")
	if dbname == "" { dbname = "nusaroute_hub" }

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: host, Port: port, User: user, Password: pass, DBName: dbname,
	})
	if err != nil {
		t.Skipf("Skipping functional test: failed to connect to DB: %v", err)
		return
	}
	defer db.Close()

    // Add basic ping test
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping DB: %v", err)
	}
}
