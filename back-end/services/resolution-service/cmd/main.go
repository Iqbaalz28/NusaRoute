package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/resolution-service/internal/model"
	"github.com/nusaroute/services/resolution-service/internal/repository"
)

func main() {
	log.Println("🚀 Starting NusaRoute Resolution Service...")
	port := getEnv("PORT", "8010")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", "nusaroute_secret"),
		DBName: getEnv("DB_NAME", "nusaroute_resolution"),
	})
	if err != nil { log.Fatalf("DB failed: %v", err) }
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	repo := repository.NewResolutionRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Kafka consumers — auto-create tickets for failure events
	cg := kafka.NewConsumerGroup()

	cg.Subscribe(ctx, kafkaBrokers, events.TopicDeliveryFailed, "resolution-service", func(ctx context.Context, key, value []byte) error {
		var evt events.DeliveryFailedEvent
		json.Unmarshal(value, &evt)
		if evt.AttemptNum >= events.MaxDeliveryAttempts {
			ticket := &model.Ticket{
				OrderID: evt.OrderID, AWB: evt.AWB, Type: events.ResolutionTypeDeliveryFailed,
				Description: "Pengiriman gagal setelah " + strconv.Itoa(evt.MaxAttempts) + " percobaan. Alasan: " + evt.Reason,
			}
			repo.CreateTicket(ctx, ticket)
			publishResolutionCreated(ctx, producer, ticket)
			log.Printf("[Resolution] Auto-created ticket for delivery failure: AWB=%s", evt.AWB)
		}
		return nil
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageLost, "resolution-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageLostSuspectedEvent
		json.Unmarshal(value, &evt)
		ticket := &model.Ticket{
			OrderID: evt.OrderID, AWB: evt.AWB, Type: events.ResolutionTypeLost,
			Description: "Paket tidak menunjukkan aktivitas selama " + strconv.Itoa(evt.HoursSinceUpdate) + " jam.",
		}
		repo.CreateTicket(ctx, ticket)
		publishResolutionCreated(ctx, producer, ticket)
		// Auto-create insurance claim
		claim := &model.Claim{TicketID: ticket.ID, OrderID: evt.OrderID, ClaimType: "INSURANCE", Amount: 0}
		repo.CreateClaim(ctx, claim)
		log.Printf("[Resolution] Auto-created LOST ticket + insurance claim: AWB=%s", evt.AWB)
		return nil
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageDamaged, "resolution-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageDamagedEvent
		json.Unmarshal(value, &evt)
		ticket := &model.Ticket{
			OrderID: evt.OrderID, AWB: evt.AWB, Type: events.ResolutionTypeDamaged,
			Description: "Kerusakan dilaporkan oleh " + evt.ReportedBy + ": " + evt.DamageDesc,
			Evidence: evt.EvidencePhotos,
		}
		repo.CreateTicket(ctx, ticket)
		publishResolutionCreated(ctx, producer, ticket)
		log.Printf("[Resolution] Auto-created DAMAGED ticket: AWB=%s", evt.AWB)
		return nil
	})

	defer cg.CloseAll()

	// HTTP endpoints
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/resolutions/tickets", func(w http.ResponseWriter, r *http.Request) {
		var req model.CreateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		userID := middleware.GetUserID(r.Context())
		ticket := &model.Ticket{OrderID: req.OrderID, AWB: req.AWB, UserID: userID, Type: req.Type, Description: req.Description}
		if err := repo.CreateTicket(r.Context(), ticket); err != nil { response.InternalError(w, err.Error()); return }
		publishResolutionCreated(r.Context(), producer, ticket)
		response.Created(w, "ticket created", ticket)
	})

	mux.HandleFunc("GET /api/v1/resolutions/tickets/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]
		ticket, err := repo.GetTicketByID(r.Context(), id)
		if err != nil { response.NotFound(w, "ticket not found"); return }
		response.Success(w, "ticket retrieved", ticket)
	})

	mux.HandleFunc("GET /api/v1/resolutions/tickets", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 { page = 1 }
		tickets, total, _ := repo.ListTickets(r.Context(), status, page, 20)
		response.Paginated(w, tickets, page, 20, total)
	})

	mux.HandleFunc("PUT /api/v1/resolutions/tickets/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]
		var req model.UpdateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		if err := repo.UpdateTicket(r.Context(), id, req); err != nil { response.InternalError(w, err.Error()); return }

		if req.Resolution != "" {
			ticket, _ := repo.GetTicketByID(r.Context(), id)
			evt := events.ResolutionResolvedEvent{
				BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionResolved,
					Timestamp: time.Now(), Source: "resolution-service"},
				TicketID: id, OrderID: ticket.OrderID, Resolution: req.Resolution,
			}
			producer.Publish(r.Context(), events.TopicResolutionResolved, id, evt)
		}
		response.Success(w, "ticket updated", nil)
	})

	mux.HandleFunc("POST /api/v1/resolutions/claims", func(w http.ResponseWriter, r *http.Request) {
		var req model.CreateClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid body"); return }
		claim := &model.Claim{TicketID: req.TicketID, OrderID: req.OrderID, ClaimType: req.ClaimType, Amount: req.Amount}
		if err := repo.CreateClaim(r.Context(), claim); err != nil { response.InternalError(w, err.Error()); return }
		response.Created(w, "claim created", claim)
	})

	mux.HandleFunc("GET /api/v1/resolutions/claims/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]
		claim, err := repo.GetClaimByID(r.Context(), id)
		if err != nil { response.NotFound(w, "claim not found"); return }
		response.Success(w, "claim retrieved", claim)
	})

	mux.HandleFunc("GET /api/v1/resolutions/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "resolution-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	log.Printf("✅ Resolution Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil { log.Fatalf("Server failed: %v", err) }
}

func publishResolutionCreated(ctx context.Context, producer *kafka.Producer, ticket *model.Ticket) {
	evt := events.ResolutionCreatedEvent{
		BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicResolutionCreated,
			Timestamp: time.Now(), Source: "resolution-service"},
		TicketID: ticket.ID, OrderID: ticket.OrderID, AWB: ticket.AWB,
		Type: ticket.Type, Priority: ticket.Priority,
	}
	producer.Publish(ctx, events.TopicResolutionCreated, ticket.ID, evt)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
