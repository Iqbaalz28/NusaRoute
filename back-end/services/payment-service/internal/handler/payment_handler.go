package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/payment-service/internal/model"
	"github.com/nusaroute/services/payment-service/internal/service"
)

type PaymentHandler struct {
	svc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/payments/initiate", h.InitiatePayment)
	mux.HandleFunc("POST /api/v1/payments/webhook", h.HandleWebhook)
	mux.HandleFunc("GET /api/v1/payments/", h.GetPaymentStatus)
	mux.HandleFunc("GET /api/v1/payments/health", h.Health)
}

func (h *PaymentHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "payment-service is healthy", map[string]string{"status": "UP"})
}

func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req model.InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	tx, err := h.svc.InitiatePayment(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "payment initiated", tx)
}

func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload model.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.BadRequest(w, "invalid webhook payload")
		return
	}

	if err := h.svc.HandleWebhook(r.Context(), payload); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "webhook processed", nil)
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		response.BadRequest(w, "order_id is required")
		return
	}
	orderID := parts[len(parts)-1]

	tx, err := h.svc.GetPaymentStatus(r.Context(), orderID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}

	response.Success(w, "payment status retrieved", tx)
}
