package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/services/resolution-service/internal/model"
)

type ResolutionRepository interface {
	CreateTicket(ctx context.Context, t *model.Ticket, outboxTopic string, outboxPayload interface{}) error
	GetTicketByID(ctx context.Context, id string) (*model.Ticket, error)
	GetTicketsByOrderID(ctx context.Context, orderID string) ([]model.Ticket, error)
	UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest, outboxTopic string, outboxPayload interface{}) error
	ListTickets(ctx context.Context, status string, page, perPage int) ([]model.Ticket, int64, error)
	ListTicketsByUser(ctx context.Context, userID string) ([]model.Ticket, error)
	CreateClaim(ctx context.Context, c *model.Claim) error
	GetClaimByID(ctx context.Context, id string) (*model.Claim, error)
	GetClaimsByOrderID(ctx context.Context, orderID string) ([]model.Claim, error)
	ListClaims(ctx context.Context, status string, page, perPage int) ([]model.Claim, int64, error)
	UpdateClaim(ctx context.Context, id, status string, amount float64) error
}

type resolutionRepo struct{ db *sqlx.DB }

func NewResolutionRepository(db *sqlx.DB) ResolutionRepository {
	return &resolutionRepo{db: db}
}

func (r *resolutionRepo) CreateTicket(ctx context.Context, t *model.Ticket, outboxTopic string, outboxPayload interface{}) error {
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

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO tickets (id, order_id, awb, user_id, type, priority, status, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err = tx.ExecContext(ctx, query,
		t.ID, t.OrderID, t.AWB, t.UserID, t.Type, t.Priority, t.Status, t.Description, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return err
	}

	if outboxTopic != "" {
		if err := outbox.InsertEvent(ctx, tx, outboxTopic, outboxPayload); err != nil {
			return err
		}
	}

	return tx.Commit()
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

func (r *resolutionRepo) UpdateTicket(ctx context.Context, id string, req model.UpdateTicketRequest, outboxTopic string, outboxPayload interface{}) error {
	now := time.Now()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if req.Status != "" && req.Resolution != "" {
		_, err = tx.ExecContext(ctx,
			"UPDATE tickets SET status = $1, resolution = $2, agent_id = $3, resolved_at = $4, updated_at = $5 WHERE id = $6",
			req.Status, req.Resolution, req.AgentID, now, now, id)
		if err != nil { return err }
	} else if req.Status != "" {
		_, err = tx.ExecContext(ctx,
			"UPDATE tickets SET status = $1, agent_id = $2, updated_at = $3 WHERE id = $4",
			req.Status, req.AgentID, now, id)
		if err != nil { return err }
	}

	if outboxTopic != "" {
		if err := outbox.InsertEvent(ctx, tx, outboxTopic, outboxPayload); err != nil {
			return err
		}
	}

	return tx.Commit()
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

func (r *resolutionRepo) ListTicketsByUser(ctx context.Context, userID string) ([]model.Ticket, error) {
	var tickets []model.Ticket
	err := r.db.SelectContext(ctx, &tickets,
		"SELECT * FROM tickets WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100", userID)
	return tickets, err
}

func (r *resolutionRepo) GetClaimsByOrderID(ctx context.Context, orderID string) ([]model.Claim, error) {
	var claims []model.Claim
	err := r.db.SelectContext(ctx, &claims,
		"SELECT * FROM claims WHERE order_id = $1 ORDER BY created_at DESC", orderID)
	return claims, err
}

func (r *resolutionRepo) ListClaims(ctx context.Context, status string, page, perPage int) ([]model.Claim, int64, error) {
	offset := (page - 1) * perPage
	var total int64
	var claims []model.Claim
	if status != "" {
		r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM claims WHERE status = $1", status)
		err := r.db.SelectContext(ctx, &claims,
			"SELECT * FROM claims WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", status, perPage, offset)
		return claims, total, err
	}
	r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM claims")
	err := r.db.SelectContext(ctx, &claims,
		"SELECT * FROM claims ORDER BY created_at DESC LIMIT $1 OFFSET $2", perPage, offset)
	return claims, total, err
}

// UpdateClaim sets a claim's status (and payout amount); stamps approved_at when
// the claim is APPROVED or PAID. The approve flag is precomputed so the status
// parameter isn't deduced as two conflicting types in one statement.
func (r *resolutionRepo) UpdateClaim(ctx context.Context, id, status string, amount float64) error {
	approve := status == "APPROVED" || status == "PAID"
	_, err := r.db.ExecContext(ctx,
		`UPDATE claims SET status = $1, amount = $2,
		    approved_at = CASE WHEN $3 THEN now() ELSE approved_at END
		 WHERE id = $4`,
		status, amount, approve, id)
	return err
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
