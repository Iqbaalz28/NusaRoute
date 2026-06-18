package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/order-service/internal/model"
	"github.com/nusaroute/services/order-service/internal/service"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/orders", h.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders", h.ListOrders)
	mux.HandleFunc("GET /api/v1/orders/all", h.ListAllOrders)
	mux.HandleFunc("PUT /api/v1/orders/status", h.UpdateStatus)
	mux.HandleFunc("GET /api/v1/orders/stats", h.GetDashboardStats)
	mux.HandleFunc("GET /api/v1/orders/volume", h.GetVolumeStats)
	mux.HandleFunc("GET /api/v1/orders/", h.GetOrder)
	mux.HandleFunc("PUT /api/v1/orders/", h.CancelOrder)
	mux.HandleFunc("GET /api/v1/orders/health", h.Health)
}

func (h *OrderHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "order-service is healthy", map[string]string{"status": "UP"})
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" { response.Unauthorized(w, "not authenticated"); return }

	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body"); return
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, req)
	if err != nil { response.BadRequest(w, err.Error()); return }
	response.Created(w, "order created", order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]
	if id == "" || id == "health" { return }

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil { response.NotFound(w, "order not found"); return }
	response.Success(w, "order retrieved", order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" { response.Unauthorized(w, "not authenticated"); return }

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 10 }

	orders, total, err := h.svc.ListOrders(r.Context(), userID, page, perPage)
	if err != nil { response.InternalError(w, err.Error()); return }
	response.Paginated(w, orders, page, perPage, total)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" { response.Unauthorized(w, "not authenticated"); return }

	parts := strings.Split(r.URL.Path, "/")
	orderID := parts[len(parts)-2]

	if err := h.svc.CancelOrder(r.Context(), orderID, userID); err != nil {
		response.BadRequest(w, err.Error()); return
	}
	response.Success(w, "order cancelled", nil)
}

func (h *OrderHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserID(r.Context()) == "" { response.Unauthorized(w, "not authenticated"); return }

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 20 }
	search := r.URL.Query().Get("search")

	orders, total, err := h.svc.ListAllOrders(r.Context(), page, perPage, search)
	if err != nil { response.InternalError(w, err.Error()); return }
	response.Paginated(w, orders, page, perPage, total)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUserID(r.Context()) == "" { response.Unauthorized(w, "not authenticated"); return }

	var req struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { response.BadRequest(w, "invalid request body"); return }
	if req.OrderID == "" || req.Status == "" { response.BadRequest(w, "order_id and status are required"); return }

	if err := h.svc.AdminUpdateStatus(r.Context(), req.OrderID, req.Status); err != nil {
		response.BadRequest(w, err.Error()); return
	}
	response.Success(w, "order status updated", map[string]string{"order_id": req.OrderID, "status": req.Status})
}

func (h *OrderHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	total, sla, err := h.svc.GetDashboardStats(r.Context())
	if err != nil {
		response.InternalError(w, err.Error()); return
	}
	response.Success(w, "stats retrieved", map[string]interface{}{
		"total_orders_today": total,
		"sla_percentage":     sla,
	})
}

func (h *OrderHandler) GetVolumeStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetVolumeStats(r.Context())
	if err != nil {
		response.InternalError(w, err.Error()); return
	}
	response.Success(w, "volume retrieved", data)
}
