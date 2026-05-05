.PHONY: all build test-unit test-functional lint clean docker-up docker-down

SERVICES = user-service payment-service pricing-service order-service \
           courier-service dispatch-service hub-service tracking-service \
           notification-service resolution-service api-gateway

# === Build ===
build-all:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc 2>/dev/null && go build -o ../../bin/$$svc ./cmd/main.go && cd ../.. || \
		(cd api-gateway && go build -o ../bin/$$svc ./cmd/main.go && cd ..); \
	done

build-%:
	@echo "Building $*..."
	@if [ -d "services/$*" ]; then \
		cd services/$* && go build -o ../../bin/$* ./cmd/main.go; \
	elif [ -d "$*" ]; then \
		cd $* && go build -o ../bin/$* ./cmd/main.go; \
	fi

# === Testing ===
test-unit:
	@echo "Running unit tests..."
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd services/$$svc 2>/dev/null && go test -tags=unit -v -count=1 ./... && cd ../.. || \
		(cd api-gateway && go test -tags=unit -v -count=1 ./... && cd ..); \
	done

test-functional:
	@echo "Running functional tests..."
	@for svc in $(SERVICES); do \
		echo "Functional testing $$svc..."; \
		cd services/$$svc 2>/dev/null && go test -tags=functional -v -count=1 ./... && cd ../.. || \
		(cd api-gateway && go test -tags=functional -v -count=1 ./... && cd ..); \
	done

test-all: test-unit test-functional

# === Lint ===
lint:
	@echo "Running go vet..."
	@for svc in $(SERVICES); do \
		echo "Vetting $$svc..."; \
		cd services/$$svc 2>/dev/null && go vet ./... && cd ../.. || \
		(cd api-gateway && go vet ./... && cd ..); \
	done

# === Docker ===
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down -v

docker-build:
	@for svc in $(SERVICES); do \
		echo "Building Docker image for $$svc..."; \
		if [ -d "services/$$svc" ]; then \
			docker build -t nusaroute/$$svc:latest -f services/$$svc/Dockerfile .; \
		elif [ -d "$$svc" ]; then \
			docker build -t nusaroute/$$svc:latest -f $$svc/Dockerfile .; \
		fi \
	done

# === Clean ===
clean:
	rm -rf bin/
	@for svc in $(SERVICES); do \
		cd services/$$svc 2>/dev/null && go clean && cd ../.. || \
		(cd api-gateway && go clean && cd ..); \
	done

# === Migrations ===
migrate-up:
	@echo "Running migrations..."
	go run scripts/migrate.go up

migrate-down:
	@echo "Reverting migrations..."
	go run scripts/migrate.go down

# === Seed ===
seed:
	@echo "Seeding data..."
	go run scripts/seed.go
