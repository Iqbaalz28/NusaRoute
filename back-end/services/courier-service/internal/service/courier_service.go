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
