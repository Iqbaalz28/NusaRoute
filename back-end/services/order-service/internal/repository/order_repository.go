package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/services/order-service/internal/model"
)

// OrderRepository defines the interface for order data access.
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetByAWB(ctx context.Context, awb string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error)
	UpdateStatus(ctx context.Context, id, status, note, createdBy string) error
	SetCourier(ctx context.Context, orderID, courierID string) error
	IncrementDeliveryAttempts(ctx context.Context, orderID string) error
	GetStuckOrders(ctx context.Context, olderThan time.Duration) ([]model.Order, error)
	GetExpiredPendingOrders(ctx context.Context, olderThan time.Duration) ([]model.Order, error)
	MarkDelivered(ctx context.Context, id string) error
	MarkCancelled(ctx context.Context, id string) error
}

// orderRepo is the PostgreSQL implementation of OrderRepository.
type orderRepo struct {
	db *sqlx.DB
}

// NewOrderRepository creates a new PostgreSQL-backed order repository.
func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepo{db: db}
}

// generateAWB generates a unique Airway Bill number with NR prefix.
func generateAWB() string {
	uid := uuid.New().String()
	// Use last 12 chars of UUID (hex) for a compact AWB
	hex := fmt.Sprintf("%s", uid[:8]+uid[9:13])
	return "NR" + hex
}

func (r *orderRepo) Create(ctx context.Context, order *model.Order) error {
	order.ID = uuid.New().String()
	order.AWB = generateAWB()
	order.Status = events.OrderStatusPendingPayment
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	query := `
		INSERT INTO orders (
			id, awb, user_id, status, service_type,
			sender_name, sender_phone, sender_address, sender_lat, sender_lng,
			receiver_name, receiver_phone, receiver_address, receiver_lat, receiver_lng,
			item_description, weight_kg, length_cm, width_cm, height_cm,
			is_insured, insured_value, shipping_cost, insurance_cost, total_cost,
			delivery_attempts, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26, $27, $28
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.AWB, order.UserID, order.Status, order.ServiceType,
		order.SenderName, order.SenderPhone, order.SenderAddress, order.SenderLat, order.SenderLng,
		order.ReceiverName, order.ReceiverPhone, order.ReceiverAddress, order.ReceiverLat, order.ReceiverLng,
		order.ItemDescription, order.WeightKg, order.LengthCm, order.WidthCm, order.HeightCm,
		order.IsInsured, order.InsuredValue, order.ShippingCost, order.InsuranceCost, order.TotalCost,
		order.DeliveryAttempts, order.CreatedAt, order.UpdatedAt,
	)
	return err
}

func (r *orderRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	err := r.db.GetContext(ctx, &order, "SELECT * FROM orders WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) GetByAWB(ctx context.Context, awb string) (*model.Order, error) {
	var order model.Order
	err := r.db.GetContext(ctx, &order, "SELECT * FROM orders WHERE awb = $1", awb)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) GetByUserID(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, 0, err
	}

	var orders []model.Order
	err = r.db.SelectContext(ctx, &orders,
		"SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id, status, note, createdBy string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the order status
	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id,
	)
	if err != nil {
		return err
	}

	// Insert status history record
	_, err = tx.ExecContext(ctx,
		`INSERT INTO order_status_history (id, order_id, status, note, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New().String(), id, status, note, createdBy, time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *orderRepo) SetCourier(ctx context.Context, orderID, courierID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET courier_id = $1, updated_at = $2 WHERE id = $3",
		courierID, time.Now(), orderID,
	)
	return err
}

func (r *orderRepo) IncrementDeliveryAttempts(ctx context.Context, orderID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET delivery_attempts = delivery_attempts + 1, updated_at = $1 WHERE id = $2",
		time.Now(), orderID,
	)
	return err
}

// GetStuckOrders finds orders in transit-related statuses that haven't been updated in `olderThan` duration.
func (r *orderRepo) GetStuckOrders(ctx context.Context, olderThan time.Duration) ([]model.Order, error) {
	cutoff := time.Now().Add(-olderThan)
	var orders []model.Order
	err := r.db.SelectContext(ctx, &orders,
		`SELECT * FROM orders 
		 WHERE status IN ($1, $2, $3) 
		   AND updated_at < $4`,
		events.OrderStatusPickedUp, events.OrderStatusInTransit, events.OrderStatusOutForDelivery,
		cutoff,
	)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetExpiredPendingOrders finds orders still pending payment beyond the given duration.
func (r *orderRepo) GetExpiredPendingOrders(ctx context.Context, olderThan time.Duration) ([]model.Order, error) {
	cutoff := time.Now().Add(-olderThan)
	var orders []model.Order
	err := r.db.SelectContext(ctx, &orders,
		`SELECT * FROM orders 
		 WHERE status = $1 
		   AND created_at < $2`,
		events.OrderStatusPendingPayment, cutoff,
	)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepo) MarkDelivered(ctx context.Context, id string) error {
	now := time.Now()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, delivered_at = $2, updated_at = $2 WHERE id = $3",
		events.OrderStatusDelivered, now, id,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO order_status_history (id, order_id, status, note, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New().String(), id, events.OrderStatusDelivered, "Package delivered successfully", "system", now,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *orderRepo) MarkCancelled(ctx context.Context, id string) error {
	now := time.Now()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, cancelled_at = $2, updated_at = $2 WHERE id = $3",
		events.OrderStatusCancelled, now, id,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO order_status_history (id, order_id, status, note, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New().String(), id, events.OrderStatusCancelled, "Order cancelled", "system", now,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
