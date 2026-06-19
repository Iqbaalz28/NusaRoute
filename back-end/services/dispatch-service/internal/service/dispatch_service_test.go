//go:build unit

package service_test

import (
	"context"
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

func (m *MockDispatchRepository) CreateOpenJob(ctx context.Context, a *model.Assignment) error {
	a.ID = "job-" + a.OrderID
	a.Status = model.AssignmentStatusOpen
	m.assignments[a.ID] = a
	return nil
}

func (m *MockDispatchRepository) ExistsLeg(ctx context.Context, orderID, leg string) (bool, error) {
	for _, a := range m.assignments {
		if a.OrderID == orderID && a.Leg == leg {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockDispatchRepository) ListOpenJobs(ctx context.Context) ([]model.Assignment, error) {
	var result []model.Assignment
	for _, a := range m.assignments {
		if a.Status == model.AssignmentStatusOpen {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *MockDispatchRepository) ListByCourier(ctx context.Context, courierID string) ([]model.Assignment, error) {
	var result []model.Assignment
	for _, a := range m.assignments {
		if a.CourierID == courierID {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *MockDispatchRepository) ClaimJob(ctx context.Context, orderID, courierID, courierName, outboxTopic string, buildPayload func(*model.Assignment) interface{}) (*model.Assignment, error) {
	for _, a := range m.assignments {
		if a.OrderID == orderID && a.Status == model.AssignmentStatusOpen {
			a.CourierID = courierID
			a.CourierName = courierName
			a.Status = "ASSIGNED"
			return a, nil
		}
	}
	return nil, nil
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
	return nil, nil
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

func (m *MockDispatchRepository) UpdateStatus(ctx context.Context, id, status string, outboxTopic string, outboxPayload interface{}) error {
	if a, ok := m.assignments[id]; ok {
		a.Status = status
	}
	return nil
}

func (m *MockDispatchRepository) MarkPickedUp(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error {
	if a, ok := m.assignments[id]; ok {
		a.Status = "PICKED_UP"
	}
	return nil
}

func (m *MockDispatchRepository) MarkCompleted(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error {
	if a, ok := m.assignments[id]; ok {
		a.Status = "COMPLETED"
	}
	return nil
}

func (m *MockDispatchRepository) CompleteActiveLeg(ctx context.Context, awb, leg string) (bool, error) {
	return false, nil
}

func TestListAssignments(t *testing.T) {
	repo := NewMockDispatchRepo()
	svc := service.NewDispatchService(repo, nil, nil, nil, nil)

	// Create a test assignment directly in the repo
	repo.assignments["a1"] = &model.Assignment{
		ID:       "a1",
		OrderID:  "order-1",
		AWB:      "AWB001",
		Status:   "ASSIGNED",
	}

	assignments, err := svc.ListAssignments(context.Background(), "", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
	if assignments[0].OrderID != "order-1" {
		t.Errorf("expected order-1, got %s", assignments[0].OrderID)
	}
}

func TestListAssignments_FilterByStatus(t *testing.T) {
	repo := NewMockDispatchRepo()
	svc := service.NewDispatchService(repo, nil, nil, nil, nil)

	repo.assignments["a1"] = &model.Assignment{ID: "a1", OrderID: "order-1", Status: "ASSIGNED"}
	repo.assignments["a2"] = &model.Assignment{ID: "a2", OrderID: "order-2", Status: "COMPLETED"}

	assignments, err := svc.ListAssignments(context.Background(), "ASSIGNED", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
}
