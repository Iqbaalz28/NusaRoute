package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/services/dispatch-service/internal/model"
)

// ErrJobTaken is returned when a courier tries to claim a job that is no longer OPEN.
var ErrJobTaken = errors.New("job already taken")

type DispatchRepository interface {
	CreateAssignment(ctx context.Context, a *model.Assignment, outboxTopic string, outboxPayload interface{}) error
	CreateOpenJob(ctx context.Context, a *model.Assignment) error
	ExistsLeg(ctx context.Context, orderID, leg string) (bool, error)
	ListOpenJobs(ctx context.Context) ([]model.Assignment, error)
	ListByCourier(ctx context.Context, courierID string) ([]model.Assignment, error)
	ClaimJob(ctx context.Context, orderID, courierID, courierName, outboxTopic string, buildPayload func(*model.Assignment) interface{}) (*model.Assignment, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Assignment, error)
	GetActiveByOrderID(ctx context.Context, orderID string) (*model.Assignment, error)
	ListAssignments(ctx context.Context, status string, page, perPage int) ([]model.Assignment, error)
	UpdateStatus(ctx context.Context, id, status string, outboxTopic string, outboxPayload interface{}) error
	MarkPickedUp(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error
	MarkCompleted(ctx context.Context, id string, outboxTopic string, outboxPayload interface{}) error
	CompleteActiveLeg(ctx context.Context, awb, leg string) (bool, error)
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

// CreateOpenJob inserts an unclaimed job (status OPEN, no courier) that couriers
// can later claim first-come-first-served.
func (r *dispatchRepo) CreateOpenJob(ctx context.Context, a *model.Assignment) error {
	a.ID = uuid.New().String()
	a.Status = model.AssignmentStatusOpen
	a.AssignedAt = time.Now()
	a.CreatedAt = time.Now()

	// Insert empty strings (not NULL) for courier fields so SELECT * scans cleanly
	// into the string-typed model; a real courier is set on claim.
	query := `INSERT INTO assignments
		(id, order_id, awb, courier_id, courier_name, status, leg,
		 pickup_lat, pickup_lng, pickup_address, dropoff_lat, dropoff_lng, dropoff_address,
		 hub_id, hub_name, assigned_at, created_at)
		VALUES ($1,$2,$3,'','',$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.OrderID, a.AWB, a.Status, a.Leg,
		a.PickupLat, a.PickupLng, a.PickupAddr, a.DropoffLat, a.DropoffLng, a.DropoffAddr,
		a.HubID, a.HubName, a.AssignedAt, a.CreatedAt)
	return err
}

func (r *dispatchRepo) ExistsLeg(ctx context.Context, orderID, leg string) (bool, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM assignments WHERE order_id = $1 AND leg = $2", orderID, leg)
	return n > 0, err
}

func (r *dispatchRepo) ListOpenJobs(ctx context.Context) ([]model.Assignment, error) {
	var jobs []model.Assignment
	err := r.db.SelectContext(ctx, &jobs,
		"SELECT * FROM assignments WHERE status = $1 ORDER BY created_at ASC", model.AssignmentStatusOpen)
	return jobs, err
}

func (r *dispatchRepo) ListByCourier(ctx context.Context, courierID string) ([]model.Assignment, error) {
	var assignments []model.Assignment
	err := r.db.SelectContext(ctx, &assignments,
		"SELECT * FROM assignments WHERE courier_id = $1 ORDER BY created_at DESC LIMIT 50", courierID)
	return assignments, err
}

// ClaimJob atomically assigns an OPEN job to a courier. Only one courier can win:
// the conditional UPDATE matches a single OPEN row. Returns ErrJobTaken if the job
// was already claimed (or doesn't exist). The CourierAssigned event is written to
// the outbox in the same transaction so the customer is notified.
func (r *dispatchRepo) ClaimJob(ctx context.Context, orderID, courierID, courierName, outboxTopic string, buildPayload func(*model.Assignment) interface{}) (*model.Assignment, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	res, err := tx.ExecContext(ctx,
		`UPDATE assignments SET courier_id = $1, courier_name = $2, status = $3, assigned_at = $4
		 WHERE order_id = $5 AND status = $6`,
		courierID, courierName, model.AssignmentStatusAssigned, now, orderID, model.AssignmentStatusOpen)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrJobTaken
	}

	var a model.Assignment
	if err := tx.GetContext(ctx, &a,
		"SELECT * FROM assignments WHERE order_id = $1 AND status = $2 ORDER BY assigned_at DESC LIMIT 1",
		orderID, model.AssignmentStatusAssigned); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJobTaken
		}
		return nil, err
	}

	if outboxTopic != "" && buildPayload != nil {
		if err := outbox.InsertEvent(ctx, tx, outboxTopic, buildPayload(&a)); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &a, nil
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

// CompleteActiveLeg marks the in-progress (ASSIGNED/PICKED_UP) assignment for the
// given AWB+leg as COMPLETED, without emitting an event — the triggering scan
// (e.g. the origin hub's inbound scan) already produced the tracking event. Keyed
// on AWB because that is what the operator actually scans (the hub console does not
// know the order_id). Returns false if nothing matched (e.g. self-dropoff has no
// first-mile job, or it was already completed), which callers treat as a no-op.
func (r *dispatchRepo) CompleteActiveLeg(ctx context.Context, awb, leg string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE assignments SET status = $1, completed_at = $2
		 WHERE awb = $3 AND leg = $4 AND status IN ($5,$6)`,
		model.AssignmentStatusCompleted, time.Now(), awb, leg,
		model.AssignmentStatusAssigned, model.AssignmentStatusPickedUp)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *dispatchRepo) GetNoShowAssignments(ctx context.Context, olderThan time.Duration) ([]model.Assignment, error) {
	threshold := time.Now().Add(-olderThan)
	var assignments []model.Assignment
	err := r.db.SelectContext(ctx, &assignments,
		"SELECT * FROM assignments WHERE status = $1 AND assigned_at < $2",
		model.AssignmentStatusAssigned, threshold)
	return assignments, err
}
