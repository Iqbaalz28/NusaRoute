#!/usr/bin/env bash
# Build & push every NusaRoute image to Azure Container Registry.
# Uses `az acr build` (server-side build in ACR) so you don't need a local
# Docker daemon or `docker push`.
#
# Usage:  ./build-push.sh <acr-name> [tag]
#   <acr-name>  ACR resource name (NOT the login server). e.g. nusarouteacrab12cd
#   [tag]       image tag, default "latest"
#
# Run from anywhere; paths are resolved relative to this script.
set -euo pipefail

ACR="${1:?usage: build-push.sh <acr-name> [tag]}"
TAG="${2:-latest}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"          # .../back-end
FRONTEND_DIR="$(cd "$BACKEND_DIR/../front-end" && pwd)" # .../front-end

# Go services built from the shared Dockerfile via SERVICE_NAME build-arg.
SERVICES=(
  user-service payment-service pricing-service order-service
  courier-service dispatch-service hub-service tracking-service
  notification-service resolution-service analytics-service
)

echo "==> Registry: $ACR   Tag: $TAG"

# api-gateway has its own Dockerfile.
echo "==> Building api-gateway"
az acr build --registry "$ACR" --image "api-gateway:$TAG" \
  --file "$BACKEND_DIR/api-gateway/Dockerfile" "$BACKEND_DIR"

for svc in "${SERVICES[@]}"; do
  echo "==> Building $svc"
  az acr build --registry "$ACR" --image "$svc:$TAG" \
    --file "$BACKEND_DIR/Dockerfile" \
    --build-arg "SERVICE_NAME=$svc" \
    "$BACKEND_DIR"
done

# Frontend — same-origin build (NEXT_PUBLIC_API_URL empty → calls /api/... relative).
echo "==> Building frontend"
az acr build --registry "$ACR" --image "frontend:$TAG" \
  --file "$FRONTEND_DIR/Dockerfile" \
  --build-arg "NEXT_PUBLIC_API_URL=" \
  "$FRONTEND_DIR"

echo "==> Done. Pushed 13 images to $ACR (tag: $TAG)."
