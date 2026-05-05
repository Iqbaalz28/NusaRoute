package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/services/resolution-service/internal/model"
)

type ResolutionRepository interface {
	CreateTicket(ctx context.Context, t *model.Ticket) error
	GetTicketByID(ctx context.Context, id string) (*model.Ticket, error)
	GetTicketsByOrderID(ctx context.Context, orderID string) ([]model.Ticket, error)
	UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest) error
	ListTickets(ctx context.Context, status string, page, perPage int) ([]model.Ticket, int64, error)
	CreateClaim(ctx context.Context, c *model.Claim) error
	GetClaimByID(ctx context.Context, id string) (*model.Claim, error)
}

type resolutionRepo struct{ db *sqlx.DB }

func NewResolutionRepository(db *sqlx.DB) ResolutionRepository {
	return &resolutionRepo{db: db}
}

func (r *resolutionRepo) CreateTicket(ctx context.Context, t *model.Ticket) error {
	t.ID = uuid.New().String()
	t.Status = events.ResolutionStatusOpen
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	// Auto-assign priority based on type
	switch t.Type {
	case events.ResolutionTypeLost:
		t.Priority = "CRITICAL"
	case events.ResolutionTypeDamaged:
		t.Priority = "HIGH"
	case events.ResolutionTypeDeliveryFailed:
		t.Priority = "MEDIUM"
	default:
		t.Priority = "LOW"
	}

	query := `INSERT INTO tickets (id, order_id, awb, user_id, type, priority, status, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.OrderID, t.AWB, t.UserID, t.Type, t.Priority, t.Status, t.Description, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *resolutionRepo) GetTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	var t model.Ticket
	err := r.db.GetContext(ctx, &t, "SELECT * FROM tickets WHERE id = $1", id)
	return &t, err
}

func (r *resolutionRepo) GetTicketsByOrderID(ctx context.Context, orderID string) ([]model.Ticket, error) {
	var tickets []model.Ticket
	err := r.db.SelectContext(ctx, &tickets,
		"SELECT * FROM tickets WHERE order_id = $1 ORDER BY created_at DESC", orderID)
	return tickets, err
}

func (r *resolutionRepo) UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest) error {
	now := time.Now()
	if req.Status != "" && req.Resolution != "" {
		_, err := r.db.ExecContext(ctx,
			"UPDATE tickets SET status = $1, resolution = $2, agent_id = $3, resolved_at = $4, updated_at = $5 WHERE id = $6",
			req.Status, req.Resolution, req.AgentID, now, now, id)
		return err
	}
	if req.Status != "" {
		_, err := r.db.ExecContext(ctx,
			"UPDATE tickets SET status = $1, agent_id = $2, updated_at = $3 WHERE id = $4",
			req.Status, req.AgentID, now, id)
		return err
	}
	return nil
}

func (r *resolutionRepo) ListTickets(ctx context.Context, status string, page, perPage int) ([]model.Ticket, int64, error) {
	offset := (page - 1) * perPage
	var total int64
	var tickets []model.Ticket

	if status != "" {
		r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM tickets WHERE status = $1", status)
		r.db.SelectContext(ctx, &tickets,
			"SELECT * FROM tickets WHERE status = $1 ORDER BY priority DESC, created_at ASC LIMIT $2 OFFSET $3",
			status, perPage, offset)
	} else {
		r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM tickets")
		r.db.SelectContext(ctx, &tickets,
			"SELECT * FROM tickets ORDER BY priority DESC, created_at ASC LIMIT $1 OFFSET $2",
			perPage, offset)
	}
	return tickets, total, nil
}

func (r *resolutionRepo) CreateClaim(ctx context.Context, c *model.Claim) error {
	c.ID = uuid.New().String()
	c.Status = "PENDING"
	c.CreatedAt = time.Now()

	query := `INSERT INTO claims (id, ticket_id, order_id, claim_type, amount, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.TicketID, c.OrderID, c.ClaimType, c.Amount, c.Status, c.CreatedAt)
	return err
}

func (r *resolutionRepo) GetClaimByID(ctx context.Context, id string) (*model.Claim, error) {
	var c model.Claim
	err := r.db.GetContext(ctx, &c, "SELECT * FROM claims WHERE id = $1", id)
	return &c, err
}
