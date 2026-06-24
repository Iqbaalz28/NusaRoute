# Multi-stage build for NusaRoute API Gateway

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy shared packages first (for caching)
COPY pkg/ ./pkg/

# Copy api-gateway files
COPY api-gateway/ ./api-gateway/

# Download dependencies
WORKDIR /app/api-gateway
RUN go mod download

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/service ./cmd/main.go

# --- Production stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /app/bin/service /usr/local/bin/service

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/service"]
