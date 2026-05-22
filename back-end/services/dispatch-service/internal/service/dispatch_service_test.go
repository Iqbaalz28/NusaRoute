//go:build unit

package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nusaroute/services/dispatch-service/internal/model"
	"github.com/nusaroute/services/dispatch-service/internal/service"
)

type MockDispatchRepository struct {
	assignments map[string]*model.Assignment
}

func NewMockDispatchRepo() *MockDispatchRepository {
	return &MockDispatchRepository{assignments: make(map[string]*model.Assignment)}
}

func (m *MockDispatchRepository) CreateAssignment(ctx context.Context, a *model.Assignment, outboxTopic string, outboxPayload interface{}) error {
	a.ID = "assign-123"
	m.assignments[a.ID] = a
	return nil
}

func (m *MockDispatchRepository) ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error) {
	var result []model.Assignment
	for _, a := range m.assignments {
		if status == "" || a.Status == status {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *MockDispatchRepository) GetNoShowAssignments(ctx context.Context, duration time.Duration) ([]model.Assignment, error) {
	return nil, nil // simplified for unit test
}

func (m *MockDispatchRepository) GetActiveByOrderID(ctx context.Context, orderID string) (*model.Assignment, error) {
	for _, a := range m.assignments {
		if a.OrderID == orderID && a.Status != "COMPLETED" && a.Status != "CANCELLED" {
			return a, nil
		}
	}
	return nil, nil
}

func (m *MockDispatchRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Assignment, error) {
	for _, a := range m.assignments {
		if a.OrderID == orderID {
			return a, nil
		}
	}
	return nil, nil
}

func (m *MockDispatchRepository) UpdateStatus(ctx context.Context, id string, status string, outboxTopic string, outboxPayload interface{}) error {
	if a, ok := m.assignments[id]; ok {
		a.Status = status
	}
	return nil
}

func TestAutoAssign_Success(t *testing.T) {
	// Mock Courier Service API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "c-1", "full_name": "Courier 1", "current_lat": -6.917, "current_lng": 107.619},
				{"id": "c-2", "full_name": "Courier 2", "current_lat": -6.920, "current_lng": 107.620},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := NewMockDispatchRepo()
	svc := service.NewDispatchService(repo, nil, server.URL)

	err := svc.AutoAssign(context.Background(), "order-123", "AWB123", -6.918, 107.618, "Jl. Merdeka")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assignments, _ := repo.ListAssignments(context.Background(), "", 1, 10)
	if len(assignments) == 0 {
		t.Fatal("expected assignment to be created")
	}

	a := assignments[0]
	if a.OrderID != "order-123" {
		t.Errorf("wrong order id: %s", a.OrderID)
	}
	// courier-1 is closer than courier-2 to pickup (-6.918, 107.618)
	if a.CourierID != "c-1" {
		t.Errorf("expected closest courier (c-1), got %s", a.CourierID)
	}
}

func TestAutoAssign_NoCourier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	repo := NewMockDispatchRepo()
	svc := service.NewDispatchService(repo, nil, server.URL)

	err := svc.AutoAssign(context.Background(), "order-123", "AWB123", -6.918, 107.618, "Jl. Merdeka")
	if err == nil {
		t.Error("expected error when no courier available")
	}
}
