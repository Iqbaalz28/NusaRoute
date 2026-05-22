import os

services = [
    {"name": "payment", "db": "nusaroute_payment", "type": "postgres"},
    {"name": "dispatch", "db": "nusaroute_dispatch", "type": "postgres"},
    {"name": "hub", "db": "nusaroute_hub", "type": "postgres"},
    {"name": "resolution", "db": "nusaroute_resolution", "type": "postgres"},
    {"name": "courier", "db": "nusaroute_courier", "type": "postgres"},
    {"name": "pricing", "db": "nusaroute_pricing", "type": "postgres"},
    {"name": "tracking", "db": "nusaroute_tracking", "type": "mongo"},
    {"name": "notification", "db": "nusaroute_notification", "type": "mongo"}, # wait, is it mongo? I need to check
]

def generate_postgres_test(svc):
    name = svc['name']
    Title = name.capitalize()
    return f"""//go:build functional

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/services/{name}-service/internal/repository"
	"github.com/nusaroute/services/{name}-service/internal/service"
)

func Test{Title}Functional(t *testing.T) {{
	host := os.Getenv("DB_HOST")
	if host == "" {{ host = "localhost" }}
	port := os.Getenv("DB_PORT")
	if port == "" {{ port = "5432" }}
	user := os.Getenv("DB_USER")
	if user == "" {{ user = "postgres" }}
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {{ pass = "postgres" }}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {{ dbname = "{svc['db']}" }}

	db, err := database.ConnectPostgres(database.PostgresConfig{{
		Host: host, Port: port, User: user, Password: pass, DBName: dbname,
	}})
	if err != nil {{
		t.Skipf("Skipping functional test: failed to connect to DB: %v", err)
		return
	}}
	defer db.Close()

    // Add basic ping test
	if err := db.Ping(); err != nil {{
		t.Fatalf("failed to ping DB: %v", err)
	}}
}}
"""

def generate_mongo_test(svc):
    name = svc['name']
    Title = name.capitalize()
    return f"""//go:build functional

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/services/{name}-service/internal/repository"
	"github.com/nusaroute/services/{name}-service/internal/service"
)

func Test{Title}Functional(t *testing.T) {{
	uri := os.Getenv("MONGO_URI")
	if uri == "" {{ uri = "mongodb://localhost:27017" }}
	dbname := os.Getenv("MONGO_DB")
	if dbname == "" {{ dbname = "{svc['db']}" }}

	client, err := database.ConnectMongo(database.MongoConfig{{
		URI: uri, Database: dbname,
	}})
	if err != nil {{
		t.Skipf("Skipping functional test: failed to connect to MongoDB: %v", err)
		return
	}}
	defer client.Disconnect(context.Background())

    // Basic test
	if err := client.Ping(context.Background(), nil); err != nil {{
		t.Fatalf("failed to ping MongoDB: %v", err)
	}}
}}
"""

base_path = "e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/services"

for svc in services:
    name = svc['name']
    # Check if we should generate
    dir_path = os.path.join(base_path, f"{name}-service", "internal", "service")
    if not os.path.exists(dir_path):
        os.makedirs(dir_path, exist_ok=True)
    
    file_path = os.path.join(dir_path, f"{name}_functional_test.go")
    if not os.path.exists(file_path):
        with open(file_path, "w", encoding='utf-8') as f:
            if svc['type'] == 'postgres':
                f.write(generate_postgres_test(svc))
            else:
                f.write(generate_mongo_test(svc))
        print(f"Generated {file_path}")
