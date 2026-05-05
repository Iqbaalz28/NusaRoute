package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/courier-service/internal/model"
	"github.com/nusaroute/services/courier-service/internal/service"
)

type CourierHandler struct{ svc service.CourierService }

func NewCourierHandler(svc service.CourierService) *CourierHandler { return &CourierHandler{svc: svc} }

func (h *CourierHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/couriers/register", h.Register)
	mux.HandleFunc("PUT /api/v1/couriers/status", h.UpdateStatus)
	mux.HandleFunc("PUT /api/v1/couriers/location", h.UpdateLocation)
	mux.HandleFunc("GET /api/v1/couriers/available", h.GetAvailable)
	mux.HandleFunc("GET /api/v1/couriers/", h.GetByID)
	mux.HandleFunc("GET /api/v1/couriers/health", h.Health)
}

func (h *CourierHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "courier-service is healthy", map[string]string{"status": "UP"})
}

func (h *CourierHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterCourierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body"); return
	}
	userID := middleware.GetUserID(r.Context())
	if userID != "" { req.UserID = userID }

	courier, err := h.svc.Register(r.Context(), req)
	if err != nil { response.BadRequest(w, err.Error()); return }
	response.Created(w, "courier registered", courier)
}

func (h *CourierHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body"); return
	}

	courierID := r.URL.Query().Get("courier_id")
	if courierID == "" { response.BadRequest(w, "courier_id required"); return }

	if err := h.svc.UpdateOnlineStatus(r.Context(), courierID, req.IsOnline); err != nil {
		response.InternalError(w, err.Error()); return
	}
	response.Success(w, "status updated", nil)
}

func (h *CourierHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body"); return
	}

	courierID := r.URL.Query().Get("courier_id")
	if courierID == "" { response.BadRequest(w, "courier_id required"); return }

	if err := h.svc.UpdateLocation(r.Context(), courierID, req.Lat, req.Lng); err != nil {
		response.InternalError(w, err.Error()); return
	}
	response.Success(w, "location updated", nil)
}

func (h *CourierHandler) GetAvailable(w http.ResponseWriter, r *http.Request) {
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	radius, _ := strconv.ParseFloat(r.URL.Query().Get("radius_km"), 64)
	if radius <= 0 { radius = 10.0 }

	couriers, err := h.svc.GetAvailableNearby(r.Context(), lat, lng, radius)
	if err != nil { response.InternalError(w, err.Error()); return }
	response.Success(w, "available couriers", couriers)
}

func (h *CourierHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]
	if id == "" || id == "health" { return }

	courier, err := h.svc.GetByID(r.Context(), id)
	if err != nil { response.NotFound(w, "courier not found"); return }
	response.Success(w, "courier retrieved", courier)
}
