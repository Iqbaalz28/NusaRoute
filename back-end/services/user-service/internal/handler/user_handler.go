package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/nusaroute/services/user-service/internal/model"
	"github.com/nusaroute/services/user-service/internal/service"
)

// UserHandler handles HTTP requests for the User Service.
type UserHandler struct {
	svc       service.UserService
	jwtSecret string
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{svc: svc, jwtSecret: jwtSecret}
}

// RegisterRoutes registers all user-related routes on the given mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public endpoints (no JWT required)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("GET /api/v1/auth/health", h.Health)

	// Protected endpoints (JWT required) — wrapped by the caller
	mux.HandleFunc("GET /api/v1/users/profile", h.GetProfile)
	mux.HandleFunc("PUT /api/v1/users/profile", h.UpdateProfile)
	mux.HandleFunc("POST /api/v1/users/addresses", h.AddAddress)
	mux.HandleFunc("GET /api/v1/users/addresses", h.ListAddresses)
	mux.HandleFunc("DELETE /api/v1/users/addresses/", h.DeleteAddress)
}

func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "user-service is healthy", map[string]string{"status": "UP"})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Public self-registration must never grant a privileged role. Force the
	// default (USER); elevated roles are assigned only via seeding/admin tooling.
	req.Role = ""

	user, err := h.svc.Register(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "user registered successfully", user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.svc.Login(r.Context(), req, h.jwtSecret)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.Success(w, "login successful", result)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}

	response.Success(w, "profile retrieved", user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "profile updated", user)
}

func (h *UserHandler) AddAddress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req model.CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	addr, err := h.svc.AddAddress(r.Context(), userID, req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "address added", addr)
}

func (h *UserHandler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	addresses, err := h.svc.ListAddresses(r.Context(), userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "addresses retrieved", addresses)
}

func (h *UserHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	// Extract address ID from path: /api/v1/users/addresses/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		response.BadRequest(w, "address ID is required")
		return
	}
	addressID := parts[len(parts)-1]

	if err := h.svc.DeleteAddress(r.Context(), addressID, userID); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "address deleted", nil)
}
