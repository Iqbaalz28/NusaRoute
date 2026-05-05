package handler

import (
	"encoding/json"
	"net/http"

	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/pricing-service/internal/model"
	"github.com/nusaroute/services/pricing-service/internal/service"
)

type PricingHandler struct {
	svc service.PricingService
}

func NewPricingHandler(svc service.PricingService) *PricingHandler {
	return &PricingHandler{svc: svc}
}

func (h *PricingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/pricing/calculate", h.Calculate)
	mux.HandleFunc("POST /api/v1/pricing/compare", h.CompareAll)
	mux.HandleFunc("GET /api/v1/pricing/services", h.GetServices)
	mux.HandleFunc("GET /api/v1/pricing/health", h.Health)
}

func (h *PricingHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "pricing-service is healthy", map[string]string{"status": "UP"})
}

func (h *PricingHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req model.PriceCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body"); return
	}
	result, err := h.svc.Calculate(r.Context(), req)
	if err != nil { response.BadRequest(w, err.Error()); return }
	response.Success(w, "price calculated", result)
}

func (h *PricingHandler) CompareAll(w http.ResponseWriter, r *http.Request) {
	var req model.PriceCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body"); return
	}
	results, err := h.svc.CalculateAll(r.Context(), req)
	if err != nil { response.InternalError(w, err.Error()); return }
	response.Success(w, "price comparison", results)
}

func (h *PricingHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.svc.GetServices(r.Context())
	if err != nil { response.InternalError(w, err.Error()); return }
	response.Success(w, "available services", services)
}
