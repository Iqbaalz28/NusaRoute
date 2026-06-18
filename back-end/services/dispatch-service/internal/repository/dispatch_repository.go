package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/services/dispatch-service/internal/model"
)

type DispatchRepository interface {
	CreateAssignment(ctx context.Context, a *model.Assignment, outboxTopic string, outboxPayload interface{}) error
	GetByOrderID(ctx context.Context, orderID string) (*model.Assignment, error)
	GetActiveByOrderID(ctx context.Context, orderID string) (*model.Assignment, error)
	ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error)
	UpdateStatus(ctx context.Context, id, status string, outboxTopic string, outboxPayload interface{}) error
	MarkPickedUp(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error
	MarkCompleted(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error
	GetNoShowAssignments(ctx context.Context, olderThan time.Duration) ([]model.Assignment, error)
}

type dispatchRepo struct{ db *sqlx.DB }

func NewDispatchRepository(db *sqlx.DB) DispatchRepository { return &dispatchRepo{db: db} }

func (r *dispatchRepo) CreateAssignment(ctx context.Context, a *model.Assignment, outboxTopic string, outboxPayload interface{}) error {
	a.ID = uuid.New().String()
	a.Status = model.AssignmentStatusAssigned
	a.AssignedAt = time.Now()
	a.CreatedAt = time.Now()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO assignments (id, order_id, awb, courier_id, courier_name, status,
		pickup_lat, pickup_lng, pickup_address, assigned_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err = tx.ExecContext(ctx, query,
		a.ID, a.OrderID, a.AWB, a.CourierID, a.CourierName, a.Status,
		a.PickupLat, a.PickupLng, a.PickupAddr, a.AssignedAt, a.CreatedAt)
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

func (r *dispatchRepo) GetByOrderID(ctx context.Context, orderID string) (*model.Assignment, error) {
	var a model.Assignment
	err := r.db.GetContext(ctx, &a, "SELECT * FROM assignments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1", orderID)
	return &a, err
}

func (r *dispatchRepo) GetActiveByOrderID(ctx context.Context, orderID string) (*model.Assignment, error) {
	var a model.Assignment
	err := r.db.GetContext(ctx, &a,
		"SELECT * FROM assignments WHERE order_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT 1",
		orderID, model.AssignmentStatusAssigned)
	return &a, err
}

func (r *dispatchRepo) ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error) {
	var assignments []model.Assignment
	offset := (page - 1) * perPage
	if status != "" {
		err := r.db.SelectContext(ctx, &assignments,
			"SELECT * FROM assignments WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			status, perPage, offset)
		return assignments, err
	}
	err := r.db.SelectContext(ctx, &assignments,
		"SELECT * FROM assignments ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		perPage, offset)
	return assignments, err
}

func (r *dispatchRepo) UpdateStatus(ctx context.Context, id, status string, outboxTopic string, outboxPayload interface{}) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE assignments SET status = $1 WHERE id = $2", status, id)
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

func (r *dispatchRepo) MarkPickedUp(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx,
		"UPDATE assignments SET status = $1, picked_up_at = $2 WHERE id = $3",
		model.AssignmentStatusPickedUp, now, id)
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

func (r *dispatchRepo) MarkCompleted(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx,
		"UPDATE assignments SET status = $1, completed_at = $2 WHERE id = $3",
		model.AssignmentStatusCompleted, now, id)
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

func (r *dispatchRepo) GetNoShowAssignments(ctx context.Context, olderThan time.Duration) ([]model.Assignment, error) {
	threshold := time.Now().Add(-olderThan)
	var assignments []model.Assignment
	err := r.db.SelectContext(ctx, &assignments,
		"SELECT * FROM assignments WHERE status = $1 AND assigned_at < $2",
		model.AssignmentStatusAssigned, threshold)
	return assignments, err
}
