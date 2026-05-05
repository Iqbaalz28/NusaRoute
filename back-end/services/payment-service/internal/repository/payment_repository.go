package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/services/payment-service/internal/model"
)

// PaymentRepository defines the interface for payment data access.
type PaymentRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	GetByOrderID(ctx context.Context, orderID string) (*model.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
	UpdateStatus(ctx context.Context, id, status string) error
	MarkPaid(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
	MarkRefunded(ctx context.Context, id string) error
}

type paymentRepo struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, tx *model.Transaction) error {
	tx.ID = uuid.New().String()
	tx.Status = model.PaymentStatusPending
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	query := `
		INSERT INTO transactions (id, order_id, amount, method, status, payment_url, external_id, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.OrderID, tx.Amount, tx.Method, tx.Status,
		tx.PaymentURL, tx.ExternalID, tx.IdempotencyKey, tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(ctx context.Context, orderID string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.GetContext(ctx, &tx, "SELECT * FROM transactions WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1", orderID)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *paymentRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.GetContext(ctx, &tx, "SELECT * FROM transactions WHERE idempotency_key = $1", key)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE transactions SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id)
	return err
}

func (r *paymentRepo) MarkPaid(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE transactions SET status = $1, paid_at = $2, updated_at = $3 WHERE id = $4",
		model.PaymentStatusPaid, now, now, id)
	return err
}

func (r *paymentRepo) MarkFailed(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE transactions SET status = $1, failed_at = $2, updated_at = $3 WHERE id = $4",
		model.PaymentStatusFailed, now, now, id)
	return err
}

func (r *paymentRepo) MarkRefunded(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE transactions SET status = $1, refunded_at = $2, updated_at = $3 WHERE id = $4",
		model.PaymentStatusRefunded, now, now, id)
	return err
}
