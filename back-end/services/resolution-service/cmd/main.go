package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nusaroute/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/kafka"
	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/resolution-service/internal/model"
	"github.com/nusaroute/services/resolution-service/internal/repository"
	"github.com/nusaroute/services/resolution-service/internal/service"
)

func main() {
	logger.InitLogger("resolution-service")
	logger.Info(context.Background(), fmt.Sprint(" Starting NusaRoute Resolution Service..."))
	port := getEnv("PORT", "8010")
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	db, err := database.ConnectPostgres(database.PostgresConfig{
		Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"),
		User: getEnv("DB_USER", "nusaroute"), Password: getEnv("DB_PASSWORD", ""),
		DBName: getEnv("DB_NAME", "nusaroute_resolution"),
	})
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("DB failed: %v", err))
	}
	defer db.Close()

	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	repo := repository.NewResolutionRepository(db)
	resolutionSvc := service.NewResolutionService(repo, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cg := kafka.NewConsumerGroup()

	cg.Subscribe(ctx, kafkaBrokers, events.TopicDeliveryFailed, "resolution-service", func(ctx context.Context, key, value []byte) error {
		var evt events.DeliveryFailedEvent
		json.Unmarshal(value, &evt)
		return resolutionSvc.AutoCreateDeliveryFailedTicket(ctx, evt.OrderID, evt.AWB, evt.Reason, evt.AttemptNum, evt.MaxAttempts)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageLost, "resolution-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageLostSuspectedEvent
		json.Unmarshal(value, &evt)
		return resolutionSvc.AutoCreateLostTicketAndClaim(ctx, evt.OrderID, evt.AWB, evt.HoursSinceUpdate)
	})

	cg.Subscribe(ctx, kafkaBrokers, events.TopicPackageDamaged, "resolution-service", func(ctx context.Context, key, value []byte) error {
		var evt events.PackageDamagedEvent
		json.Unmarshal(value, &evt)
		return resolutionSvc.AutoCreateDamagedTicket(ctx, evt.OrderID, evt.AWB, evt.ReportedBy, evt.DamageDesc, evt.EvidencePhotos)
	})

	defer cg.CloseAll()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// --- Customer endpoints (JWT, any authenticated user) ---
	mux.HandleFunc("POST /api/v1/resolutions/tickets", func(w http.ResponseWriter, r *http.Request) {
		var req model.CreateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		userID := middleware.GetUserID(r.Context())
		ticket := &model.Ticket{OrderID: req.OrderID, AWB: req.AWB, UserID: userID, Type: req.Type, Description: req.Description}
		if err := resolutionSvc.CreateTicket(r.Context(), ticket); err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Created(w, "ticket created", ticket)
	})

	// All tickets + claims for one order (powers the customer order-detail panel).
	mux.HandleFunc("GET /api/v1/resolutions/order/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		tickets, claims, err := resolutionSvc.GetOrderResolution(r.Context(), r.PathValue("orderId"))
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "order resolution retrieved", map[string]interface{}{"tickets": tickets, "claims": claims})
	})

	// Tickets owned by the authenticated user.
	mux.HandleFunc("GET /api/v1/resolutions/my-tickets", func(w http.ResponseWriter, r *http.Request) {
		tickets, err := resolutionSvc.ListUserTickets(r.Context(), middleware.GetUserID(r.Context()))
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "my tickets retrieved", tickets)
	})

	mux.HandleFunc("GET /api/v1/resolutions/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
		ticket, err := resolutionSvc.GetTicketByID(r.Context(), r.PathValue("id"))
		if err != nil {
			response.NotFound(w, "ticket not found")
			return
		}
		response.Success(w, "ticket retrieved", ticket)
	})

	mux.HandleFunc("POST /api/v1/resolutions/claims", func(w http.ResponseWriter, r *http.Request) {
		var req model.CreateClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		claim := &model.Claim{TicketID: req.TicketID, OrderID: req.OrderID, ClaimType: req.ClaimType, Amount: req.Amount}
		if err := resolutionSvc.CreateClaim(r.Context(), claim); err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Created(w, "claim created", claim)
	})

	mux.HandleFunc("GET /api/v1/resolutions/claims/{id}", func(w http.ResponseWriter, r *http.Request) {
		claim, err := resolutionSvc.GetClaimByID(r.Context(), r.PathValue("id"))
		if err != nil {
			response.NotFound(w, "claim not found")
			return
		}
		response.Success(w, "claim retrieved", claim)
	})

	// --- Admin endpoints (gateway gates /api/v1/resolutions/admin/ to ADMIN) ---
	mux.HandleFunc("GET /api/v1/resolutions/admin/tickets", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		tickets, total, _ := resolutionSvc.ListTickets(r.Context(), status, page, 50)
		response.Paginated(w, tickets, page, 50, total)
	})

	mux.HandleFunc("PUT /api/v1/resolutions/admin/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
		var req model.UpdateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		if req.AgentID == "" {
			req.AgentID = middleware.GetUserID(r.Context())
		}
		if err := resolutionSvc.UpdateTicket(r.Context(), r.PathValue("id"), req); err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "ticket updated", nil)
	})

	mux.HandleFunc("GET /api/v1/resolutions/admin/claims", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		claims, total, _ := resolutionSvc.ListClaims(r.Context(), status, page, 50)
		response.Paginated(w, claims, page, 50, total)
	})

	mux.HandleFunc("PUT /api/v1/resolutions/admin/claims/{id}", func(w http.ResponseWriter, r *http.Request) {
		var req model.UpdateClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid body")
			return
		}
		if err := resolutionSvc.UpdateClaim(r.Context(), r.PathValue("id"), req.Status, req.Amount); err != nil {
			response.InternalError(w, err.Error())
			return
		}
		response.Success(w, "claim updated", nil)
	})

	mux.HandleFunc("GET /api/v1/resolutions/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "resolution-service is healthy", map[string]string{"status": "UP"})
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.HeaderAuth(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	h = middleware.Metrics(h)

	logger.Info(context.Background(), fmt.Sprintf(" Resolution Service running on port %s", port))
	if err := http.ListenAndServe(":"+port, h); err != nil {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
