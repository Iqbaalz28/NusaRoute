//go:build unit

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/services/resolution-service/internal/model"
	"github.com/nusaroute/services/resolution-service/internal/service"
)

type MockResolutionRepository struct {
	tickets map[string]*model.Ticket
	claims  map[string]*model.Claim
}

func NewMockResolutionRepo() *MockResolutionRepository {
	return &MockResolutionRepository{
		tickets: make(map[string]*model.Ticket),
		claims:  make(map[string]*model.Claim),
	}
}

func (m *MockResolutionRepository) CreateTicket(ctx context.Context, ticket *model.Ticket, outboxTopic string, outboxPayload interface{}) error {
	ticket.ID = "ticket-123"
	m.tickets[ticket.ID] = ticket
	return nil
}

func (m *MockResolutionRepository) GetTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	if t, ok := m.tickets[id]; ok {
		return t, nil
	}
	return nil, errors.New("not found")
}

func (m *MockResolutionRepository) GetTicketsByOrderID(ctx context.Context, orderID string) ([]model.Ticket, error) {
	var res []model.Ticket
	for _, t := range m.tickets {
		if t.OrderID == orderID {
			res = append(res, *t)
		}
	}
	return res, nil
}

func (m *MockResolutionRepository) ListTickets(ctx context.Context, status string, page, perPage int) ([]model.Ticket, int64, error) {
	var res []model.Ticket
	for _, t := range m.tickets {
		if status == "" || t.Status == status {
			res = append(res, *t)
		}
	}
	return res, int64(len(res)), nil
}

func (m *MockResolutionRepository) UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest, outboxTopic string, outboxPayload interface{}) error {
	if t, ok := m.tickets[id]; ok {
		t.Status = req.Status
		t.Resolution = req.Resolution
		return nil
	}
	return errors.New("not found")
}

func (m *MockResolutionRepository) CreateClaim(ctx context.Context, claim *model.Claim) error {
	claim.ID = "claim-123"
	m.claims[claim.ID] = claim
	return nil
}

func (m *MockResolutionRepository) GetClaimByID(ctx context.Context, id string) (*model.Claim, error) {
	if c, ok := m.claims[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func TestAutoCreateDeliveryFailedTicket(t *testing.T) {
	repo := NewMockResolutionRepo()
	svc := service.NewResolutionService(repo, nil)

	// Should not create if attempt < max
	err := svc.AutoCreateDeliveryFailedTicket(context.Background(), "ord-1", "AWB1", "Nobody home", 1, 3)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if len(repo.tickets) != 0 { t.Error("should not create ticket yet") }

	// Should create if attempt >= max
	err = svc.AutoCreateDeliveryFailedTicket(context.Background(), "ord-1", "AWB1", "Nobody home", 3, 3)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if len(repo.tickets) != 1 { t.Error("should create ticket") }
	
	for _, tk := range repo.tickets {
		if tk.Type != events.ResolutionTypeDeliveryFailed {
			t.Errorf("expected type %s, got %s", events.ResolutionTypeDeliveryFailed, tk.Type)
		}
	}
}

func TestAutoCreateLostTicketAndClaim(t *testing.T) {
	repo := NewMockResolutionRepo()
	svc := service.NewResolutionService(repo, nil)

	err := svc.AutoCreateLostTicketAndClaim(context.Background(), "ord-2", "AWB2", 48)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	
	if len(repo.tickets) != 1 { t.Error("should create ticket") }
	if len(repo.claims) != 1 { t.Error("should create claim") }

	for _, c := range repo.claims {
		if c.ClaimType != "INSURANCE" {
			t.Errorf("expected INSURANCE, got %s", c.ClaimType)
		}
	}
}
