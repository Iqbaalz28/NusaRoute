//go:build unit

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nusaroute/services/courier-service/internal/model"
	"github.com/nusaroute/services/courier-service/internal/service"
)

type MockCourierRepository struct {
	couriers map[string]*model.Courier
}

func NewMockCourierRepo() *MockCourierRepository {
	return &MockCourierRepository{couriers: make(map[string]*model.Courier)}
}

func (m *MockCourierRepository) GetByUserID(ctx context.Context, userID string) (*model.Courier, error) {
	for _, c := range m.couriers {
		if c.UserID == userID {
			return c, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockCourierRepository) Create(ctx context.Context, c *model.Courier) error {
	c.ID = "courier-123"
	m.couriers[c.ID] = c
	return nil
}

func (m *MockCourierRepository) GetByID(ctx context.Context, id string) (*model.Courier, error) {
	if c, ok := m.couriers[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (m *MockCourierRepository) GetAvailable(ctx context.Context, lat, lng, radiusKm float64) ([]model.Courier, error) {
	var available []model.Courier
	for _, c := range m.couriers {
		if c.IsOnline && c.IsAvailable {
			available = append(available, *c)
		}
	}
	return available, nil
}

func (m *MockCourierRepository) UpdateStatus(ctx context.Context, id string, isOnline bool) error {
	if c, ok := m.couriers[id]; ok {
		c.IsOnline = isOnline
		return nil
	}
	return errors.New("not found")
}

func (m *MockCourierRepository) UpdateLocation(ctx context.Context, id string, lat, lng float64) error {
	if c, ok := m.couriers[id]; ok {
		c.CurrentLat = lat
		c.CurrentLng = lng
		return nil
	}
	return errors.New("not found")
}

func (m *MockCourierRepository) SetAvailability(ctx context.Context, id string, available bool) error {
	if c, ok := m.couriers[id]; ok {
		c.IsAvailable = available
		return nil
	}
	return errors.New("not found")
}

func (m *MockCourierRepository) IncrementDeliveries(ctx context.Context, id string) error {
	if c, ok := m.couriers[id]; ok {
		c.TotalDeliveries++
		return nil
	}
	return errors.New("not found")
}

func TestRegisterCourier_Success(t *testing.T) {
	repo := NewMockCourierRepo()
	svc := service.NewCourierService(repo)

	courier, err := svc.Register(context.Background(), model.RegisterCourierRequest{
		UserID: "user-1", FullName: "Budi Kurir", Phone: "0812345678",
		VehicleType: "MOTORCYCLE", VehiclePlate: "D 1234 AB",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if courier.ID == "" {
		t.Error("expected non-empty courier ID")
	}
}

func TestRegisterCourier_MissingFields(t *testing.T) {
	repo := NewMockCourierRepo()
	svc := service.NewCourierService(repo)

	_, err := svc.Register(context.Background(), model.RegisterCourierRequest{
		UserID: "user-1", Phone: "0812345678",
	})

	if err == nil {
		t.Error("expected error for missing full name")
	}
}

func TestUpdateOnlineStatus(t *testing.T) {
	repo := NewMockCourierRepo()
	svc := service.NewCourierService(repo)

	c, _ := svc.Register(context.Background(), model.RegisterCourierRequest{
		UserID: "user-1", FullName: "Budi Kurir", Phone: "0812345678",
	})

	err := svc.UpdateOnlineStatus(context.Background(), c.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := svc.GetByID(context.Background(), c.ID)
	if !updated.IsOnline {
		t.Error("expected courier to be online")
	}
}
