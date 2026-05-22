//go:build functional

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/nusaroute/pkg/database"
)

func TestTrackingFunctional(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" { uri = "mongodb://localhost:27017" }
	dbname := os.Getenv("MONGO_DB")
	if dbname == "" { dbname = "nusaroute_tracking" }

	client, err := database.ConnectMongo(database.MongoConfig{
		URI: uri, DBName: dbname,
	})
	if err != nil {
		t.Skipf("Skipping functional test: failed to connect to MongoDB: %v", err)
		return
	}
	defer client.Client().Disconnect(context.Background())

    // Basic test
	if err := client.Client().Ping(context.Background(), nil); err != nil {
		t.Fatalf("failed to ping MongoDB: %v", err)
	}
}
