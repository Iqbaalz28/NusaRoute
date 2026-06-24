#!/usr/bin/env bash
# Bootstrap the database schema + seed data on a freshly deployed cluster.
# The services do NOT auto-apply migrations (runMigrations is a no-op), so this
# script applies every service's migrations to its database, loads the data
# warehouse, and (optionally) runs the Go seeder for demo hubs/orders.
#
# Prereqs: kubectl context points at the AKS cluster; the postgres pod is up;
#          `go` on PATH if you want the demo seeder to run.
# Usage:   ./db-bootstrap.sh
set -euo pipefail

NS=nusaroute
POD=postgres-0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"   # .../back-end

echo "==> Waiting for $POD to be ready..."
kubectl wait --for=condition=ready "pod/$POD" -n "$NS" --timeout=300s

# Apply one .sql file (stdin) to a given database, failing fast on error.
apply() { # $1=db  $2=file
  echo "    - $(basename "$2")  ->  $1"
  kubectl exec -i -n "$NS" "$POD" -- \
    sh -c "PGPASSWORD=\$POSTGRES_PASSWORD psql -U nusaroute -v ON_ERROR_STOP=1 -d $1" < "$2"
}

# service-dir : database  (Mongo services tracking/notification have no SQL)
PAIRS=(
  "user-service:nusaroute_user"
  "payment-service:nusaroute_payment"
  "pricing-service:nusaroute_pricing"
  "order-service:nusaroute_order"
  "courier-service:nusaroute_courier"
  "dispatch-service:nusaroute_dispatch"
  "hub-service:nusaroute_hub"
  "resolution-service:nusaroute_resolution"
)

echo "==> Applying service migrations..."
for pair in "${PAIRS[@]}"; do
  svc="${pair%%:*}"; db="${pair##*:}"
  dir="$BACKEND_DIR/services/$svc/migrations"
  [ -d "$dir" ] || { echo "    (no migrations dir for $svc, skipping)"; continue; }
  for f in $(ls "$dir"/*.sql 2>/dev/null | sort); do
    apply "$db" "$f"
  done
done

echo "==> Loading data warehouse (star schema + seed + ML inputs)..."
apply "nusaroute_warehouse" "$BACKEND_DIR/data-warehouse/data_warehouse.init.sql"

echo "==> Schema + warehouse ready."

# --- Optional demo seeder (hubs + orders) ----------------------------------
if command -v go >/dev/null 2>&1; then
  echo "==> Seeding demo hubs/orders via port-forward..."
  PGPW="$(kubectl get secret -n "$NS" nusaroute-secrets -o jsonpath='{.data.db-password}' | base64 -d)"
  kubectl port-forward -n "$NS" svc/postgres 5433:5432 >/dev/null 2>&1 &
  PF=$!
  # give the forward a moment; trap ensures we always kill it
  trap 'kill $PF >/dev/null 2>&1 || true' EXIT
  for i in $(seq 1 10); do
    kubectl exec -n "$NS" "$POD" -- true >/dev/null 2>&1 && break || sleep 1
  done
  sleep 2
  ( cd "$BACKEND_DIR" && DB_HOST=localhost DB_PORT=5433 DB_USER=nusaroute DB_PASSWORD="$PGPW" \
      go run ./scripts/seed.go ) || echo "    (seeder failed — run it manually; see AZURE_DEPLOY.md)"
else
  echo "==> 'go' not found; skipping demo seeder."
  echo "    To seed later: kubectl port-forward -n $NS svc/postgres 5433:5432"
  echo "    then: DB_HOST=localhost DB_PORT=5433 DB_PASSWORD=<db-password> go run ./scripts/seed.go"
fi

echo "==> Done."
