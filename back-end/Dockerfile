# Multi-stage build for NusaRoute Go microservices
# Usage: docker build --build-arg SERVICE_NAME=<service-name> -t nusaroute/<service-name> .

FROM golang:1.25-alpine AS builder

ARG SERVICE_NAME

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy shared packages first (for caching)
COPY pkg/ ./pkg/

# Copy all services (needed for cross-service replace directives in go.mod)
COPY services/ ./services/

# Copy API Gateway
COPY api-gateway/ ./api-gateway/

# Download dependencies
WORKDIR /app/services/${SERVICE_NAME}
RUN go mod download

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/service ./cmd/main.go

# Copy migrations if they exist (will be empty dir if not present)
RUN mkdir -p /app/migrations && \
    if [ -d "migrations" ]; then cp -r migrations/* /app/migrations/ 2>/dev/null || true; fi

# --- Production stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /app/bin/service /usr/local/bin/service
COPY --from=builder /app/migrations/ /migrations/

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/service"]
