# 🚚 NusaRoute — Microservices Logistics Platform

> Tugas Besar Mata Kuliah **Distributed, Parallel & Cloud Computing** — Universitas Pendidikan Indonesia, Semester 4

NusaRoute adalah platform logistik berbasis **arsitektur microservices** yang dibangun menggunakan **Go 1.23**, **Apache Kafka** (Event-Driven Architecture), dan **database-per-service pattern**.

---

## 📐 Arsitektur

```
┌──────────────┐     ┌──────────────────────────────────────────────────────┐
│  Front-End   │────▶│                  API Gateway (:8080)                 │
│  Dashboard   │     │         JWT Validation · Rate Limiting · CORS       │
└──────────────┘     └──────────┬───────┬───────┬───────┬───────┬──────────┘
                               │       │       │       │       │
         ┌─────────────────────┼───────┼───────┼───────┼───────┼────────────────┐
         │                     ▼       ▼       ▼       ▼       ▼                │
         │  ┌──────────┐ ┌─────────┐ ┌───────┐ ┌────────┐ ┌──────────┐         │
         │  │  User    │ │ Payment │ │ Order │ │Pricing │ │ Courier  │         │
         │  │ Service  │ │ Service │ │Service│ │Service │ │ Service  │         │
         │  │  :8001   │ │  :8002  │ │ :8004 │ │ :8003  │ │  :8005   │         │
         │  └──────────┘ └─────────┘ └───────┘ └────────┘ └──────────┘         │
         │  ┌──────────┐ ┌─────────┐ ┌───────────┐ ┌────────────┐             │
         │  │ Dispatch │ │   Hub   │ │ Tracking  │ │Notification│             │
         │  │ Service  │ │ Service │ │  Service  │ │  Service   │             │
         │  │  :8006   │ │  :8007  │ │   :8008   │ │   :8009    │             │
         │  └──────────┘ └─────────┘ └───────────┘ └────────────┘             │
         │  ┌────────────┐                                                     │
         │  │ Resolution │        ◄══════ Apache Kafka ══════►                 │
         │  │  Service   │            Event-Driven Messaging                   │
         │  │   :8010    │                                                     │
         │  └────────────┘                                                     │
         └─────────────────────────────────────────────────────────────────────┘
```

---

## 🛠 Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.23 |
| Messaging | Apache Kafka + Zookeeper |
| Relational DB | PostgreSQL 16 (+ PostGIS for Dispatch) |
| Document DB | MongoDB 7 (Tracking, Notification) |
| Cache | Redis 7 (JWT blacklist, price cache, GPS) |
| Object Storage | MinIO (S3-compatible, evidence photos) |
| API Gateway | Custom Go reverse proxy |
| CI/CD | Jenkins (8-stage pipeline) |
| Container | Docker & Docker Compose |
| Orchestration | Kubernetes (HPA autoscaling) |
| Frontend | Vanilla HTML/CSS/JS (Glassmorphism UI) |

---

## 📦 Microservices

| # | Service | Port | Database | Key Features |
|---|---------|------|----------|-------------|
| 1 | User Service | 8001 | PostgreSQL + Redis | JWT auth, bcrypt, address book |
| 2 | Payment Service | 8002 | PostgreSQL | Idempotency key, webhook handler |
| 3 | Pricing Service | 8003 | PostgreSQL + Redis | Haversine distance, volumetric weight |
| 4 | Order Service | 8004 | PostgreSQL | AWB generation, Saga pattern, SLA monitor |
| 5 | Courier Service | 8005 | PostgreSQL | Availability tracking, nearby search |
| 6 | Dispatch Service | 8006 | PostgreSQL | VRP nearest-courier, no-show monitor |
| 7 | Hub Service | 8007 | PostgreSQL | 3-type scan (inbound/sort/outbound) |
| 8 | Tracking Service | 8008 | MongoDB + Redis | Immutable event ledger, GPS tracking |
| 9 | Notification Service | 8009 | MongoDB | Multi-channel (Email/WA/Push) |
| 10 | Resolution Service | 8010 | PostgreSQL | Auto-ticket, insurance claims, returns |

---

## 🔄 Event-Driven Flow (Kafka Topics)

```
payment.success          → Order Service
order.ready-for-pickup   → Dispatch Service, Notification Service
courier.assigned         → Tracking Service, Notification Service
package.picked-up        → Tracking Service, Notification Service
package.scanned-at-hub   → Tracking Service
package.delivered        → Order Service, Tracking Service, Notification Service
delivery.failed          → Resolution Service, Notification Service, Order Service
package.lost.suspected   → Resolution Service, Notification Service
package.damaged          → Resolution Service, Notification Service
order.cancelled          → Payment Service (refund), Notification Service
resolution.created       → Notification Service
resolution.resolved      → Order Service, Notification Service
```

---

## ⚠️ Fault Handling

| Skenario | Solusi |
|----------|--------|
| Paket tidak sampai | 3x retry → RETURN_TO_SENDER |
| Paket hilang | SLA Monitor (48h stuck) → auto insurance claim |
| Paket rusak | PACKAGE_DAMAGED event → auto-ticket |
| Payment timeout | Saga: auto-cancel setelah 24h |
| Kurir no-show | 2h timeout → auto-reassign |
| Event gagal diproses | Dead Letter Queue (DLQ) setelah 3x retry |

---

## 🚀 Quick Start

### Prerequisites
- Go 1.23+
- Docker & Docker Compose

### 1. Start Infrastructure
```bash
docker-compose -f back-end/docker-compose.yml up -d postgres mongodb redis kafka zookeeper minio kafka-ui
```

### 2. Run Services Individually
```bash
cd back-end/services/user-service && go run cmd/main.go
cd back-end/services/order-service && go run cmd/main.go
# ... etc
```

### 3. Or Run All via Docker Compose
```bash
cd back-end && docker-compose up --build
```

### 4. Access
| Service | URL |
|---------|-----|
| API Gateway | http://localhost:8080 |
| Frontend Dashboard | Open `front-end/index.html` |
| Kafka UI | http://localhost:8090 |
| MinIO Console | http://localhost:9001 |

---

## 🧪 Testing

```bash
# Unit tests (no DB required)
cd back-end && go test -tags=unit -v ./...

# Build all services
cd back-end && make build-all
```

---

## 📁 Project Structure

```
NusaRoute/
├── front-end/                   # Web Dashboard (HTML/CSS/JS)
│   ├── index.html
│   ├── index.css
│   └── app.js
├── back-end/
│   ├── docker-compose.yml       # Full infrastructure stack
│   ├── Makefile                 # Build, test, lint commands
│   ├── Dockerfile               # Generic multi-stage build
│   ├── Jenkinsfile              # 8-stage CI/CD pipeline
│   ├── go.work                  # Go workspace (mono-repo)
│   ├── pkg/                     # Shared libraries
│   │   ├── events/              # Event schemas & Kafka topics
│   │   ├── kafka/               # Producer/Consumer wrapper
│   │   ├── middleware/          # JWT, Logging, Recovery, CORS
│   │   ├── database/           # PostgreSQL, MongoDB, Redis helpers
│   │   └── response/           # Standard API response format
│   ├── api-gateway/             # Custom reverse proxy
│   ├── services/                # 10 microservices
│   │   ├── user-service/
│   │   ├── payment-service/
│   │   ├── pricing-service/
│   │   ├── order-service/
│   │   ├── courier-service/
│   │   ├── dispatch-service/
│   │   ├── hub-service/
│   │   ├── tracking-service/
│   │   ├── notification-service/
│   │   └── resolution-service/
│   ├── deployments/k8s/         # Kubernetes manifests
│   └── scripts/                 # Database init scripts
└── README.md
```

---

## 👥 Tim

Tugas Besar — Distributed, Parallel & Cloud Computing  
Universitas Pendidikan Indonesia · Semester 4 · 2026

---

## 📄 License

This project is for educational purposes.
