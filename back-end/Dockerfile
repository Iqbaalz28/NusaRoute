# Multi-stage build for NusaRoute Go microservices
# Usage: docker build -t nusaroute/<service-name> -f services/<service-name>/Dockerfile .

FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy shared packages first (for caching)
COPY pkg/ ./pkg/

# Copy service-specific files
ARG SERVICE_NAME
COPY services/${SERVICE_NAME}/ ./services/${SERVICE_NAME}/

# Download dependencies
WORKDIR /app/services/${SERVICE_NAME}
RUN go mod download

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/service ./cmd/main.go

# --- Production stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=builder /app/bin/service /usr/local/bin/service
COPY --from=builder /app/services/${SERVICE_NAME}/migrations/ /migrations/

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/service"]
