package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nusaroute/services/courier-service/internal/model"
	"github.com/nusaroute/services/courier-service/internal/repository"
)

type CourierService interface {
	Register(ctx context.Context, req model.RegisterCourierRequest) (*model.Courier, error)
	Ensure(ctx context.Context, userID string, req model.EnsureCourierRequest) (*model.Courier, error)
	GetByID(ctx context.Context, id string) (*model.Courier, error)
	GetAvailableNearby(ctx context.Context, lat, lng, radiusKm float64) ([]model.Courier, error)
	UpdateOnlineStatus(ctx context.Context, courierID string, isOnline bool) error
	UpdateLocation(ctx context.Context, courierID string, lat, lng float64) error
	SetAvailability(ctx context.Context, courierID string, available bool) error
	GetDashboardStats(ctx context.Context) (int64, error)
}

type courierService struct{ repo repository.CourierRepository }

func NewCourierService(repo repository.CourierRepository) CourierService {
	return &courierService{repo: repo}
}

func (s *courierService) Register(ctx context.Context, req model.RegisterCourierRequest) (*model.Courier, error) {
	if req.FullName == "" || req.Phone == "" {
		return nil, errors.New("full_name and phone are required")
	}

	existing, _ := s.repo.GetByUserID(ctx, req.UserID)
	if existing != nil {
		return nil, errors.New("courier already registered for this user")
	}

	courier := &model.Courier{
		UserID: req.UserID, FullName: req.FullName, Phone: req.Phone, Email: req.Email,
		VehicleType: req.VehicleType, VehiclePlate: req.VehiclePlate,
		MaxCapacityKg: req.MaxCapacityKg, CoverageArea: req.CoverageArea,
	}

	if err := s.repo.Create(ctx, courier); err != nil {
		return nil, fmt.Errorf("failed to register courier: %w", err)
	}
	return courier, nil
}

// Ensure get-or-creates a courier row for a user and marks them online. Called
// when a COURIER opens their console, so accounts are auto-linked to courier data.
func (s *courierService) Ensure(ctx context.Context, userID string, req model.EnsureCourierRequest) (*model.Courier, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	// GetByUserID returns a non-nil pointer even on not-found, so gate on the
	// error (and a real ID) before treating the courier as already existing.
	if existing, err := s.repo.GetByUserID(ctx, userID); err == nil && existing != nil && existing.ID != "" {
		_ = s.repo.UpdateStatus(ctx, existing.ID, true) // mark online
		existing.IsOnline = true
		return existing, nil
	}

	name := req.FullName
	if name == "" {
		name = "Kurir"
	}
	phone := req.Phone
	if phone == "" {
		phone = "-"
	}
	c := &model.Courier{
		UserID: userID, FullName: name, Phone: phone, Email: req.Email,
		VehicleType: "MOTORCYCLE",
		CurrentLat:  -6.2, CurrentLng: 106.8, // default to Jakarta hub area
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("failed to create courier: %w", err)
	}
	_ = s.repo.UpdateStatus(ctx, c.ID, true)
	c.IsOnline = true
	return c, nil
}

func (s *courierService) GetByID(ctx context.Context, id string) (*model.Courier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *courierService) GetAvailableNearby(ctx context.Context, lat, lng, radiusKm float64) ([]model.Courier, error) {
	if radiusKm <= 0 { radiusKm = 10.0 } // default 10km radius
	return s.repo.GetAvailable(ctx, lat, lng, radiusKm)
}

func (s *courierService) UpdateOnlineStatus(ctx context.Context, courierID string, isOnline bool) error {
	return s.repo.UpdateStatus(ctx, courierID, isOnline)
}

func (s *courierService) UpdateLocation(ctx context.Context, courierID string, lat, lng float64) error {
	return s.repo.UpdateLocation(ctx, courierID, lat, lng)
}

func (s *courierService) SetAvailability(ctx context.Context, courierID string, available bool) error {
	return s.repo.SetAvailability(ctx, courierID, available)
}

func (s *courierService) GetDashboardStats(ctx context.Context) (int64, error) {
	return s.repo.GetDashboardStats(ctx)
}
