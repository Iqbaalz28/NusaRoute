package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/pkg/events"
	"github.com/nusaroute/pkg/outbox"
	"github.com/nusaroute/services/order-service/internal/model"
)

// OrderRepository defines the interface for order data access.
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetByAWB(ctx context.Context, awb string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID string, page, perPage int) ([]model.Order, int64, error)
	ListAll(ctx context.Context, page, perPage int, search string) ([]model.Order, int64, error)
	UpdateStatus(ctx context.Context, id, status, note, createdBy, outboxTopic string, outboxPayload interface{}) error
	SetCourier(ctx context.Context, orderID, courierID string) error
	IncrementDeliveryAttempts(ctx context.Context, orderID string) error
	GetStuckOrders(ctx context.Context, olderThan time.Duration) ([]model.Order, error)
	GetExpiredPendingOrders(ctx context.Context, olderThan time.Duration) ([]model.Order, error)
	MarkDelivered(ctx context.Context, id string) error
	MarkCancelled(ctx context.Context, id, outboxTopic string, outboxPayload interface{}) error
	GetDashboardStats(ctx context.Context) (totalOrdersToday int64, slaPercentage float64, err error)
	GetVolumeStats(ctx context.Context) ([]model.DailyVolume, error)
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

	if order.DeliveryMode == "" {
		order.DeliveryMode = "VIA_HUB"
	}
	query := `
		INSERT INTO orders (
			id, awb, user_id, status, service_type,
			sender_name, sender_phone, sender_address, sender_lat, sender_lng,
			receiver_name, receiver_phone, receiver_address, receiver_lat, receiver_lng,
			item_description, weight_kg, length_cm, width_cm, height_cm,
			is_insured, insured_value, shipping_cost, insurance_cost, total_cost,
			delivery_attempts, created_at, updated_at, delivery_mode,
			origin_hub_code, origin_hub_name, dest_hub_code, dest_hub_name,
			pickup_mode, dest_hub_id, sender_city, receiver_city
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26, $27, $28, $29,
			$30, $31, $32, $33,
			$34, $35, $36, $37
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.AWB, order.UserID, order.Status, order.ServiceType,
		order.SenderName, order.SenderPhone, order.SenderAddress, order.SenderLat, order.SenderLng,
		order.ReceiverName, order.ReceiverPhone, order.ReceiverAddress, order.ReceiverLat, order.ReceiverLng,
		order.ItemDescription, order.WeightKg, order.LengthCm, order.WidthCm, order.HeightCm,
		order.IsInsured, order.InsuredValue, order.ShippingCost, order.InsuranceCost, order.TotalCost,
		order.DeliveryAttempts, order.CreatedAt, order.UpdatedAt, order.DeliveryMode,
		order.OriginHubCode, order.OriginHubName, order.DestHubCode, order.DestHubName,
		order.PickupMode, order.DestHubID, order.SenderCity, order.ReceiverCity,
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

// ListAll returns every order (admin view), newest first, with optional
// case-insensitive search across AWB / receiver / sender name.
func (r *orderRepo) ListAll(ctx context.Context, page, perPage int, search string) ([]model.Order, int64, error) {
	offset := (page - 1) * perPage
	var total int64
	var orders []model.Order

	if search != "" {
		like := "%" + search + "%"
		if err := r.db.GetContext(ctx, &total,
			"SELECT COUNT(*) FROM orders WHERE awb ILIKE $1 OR receiver_name ILIKE $1 OR sender_name ILIKE $1", like); err != nil {
			return nil, 0, err
		}
		err := r.db.SelectContext(ctx, &orders,
			"SELECT * FROM orders WHERE awb ILIKE $1 OR receiver_name ILIKE $1 OR sender_name ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			like, perPage, offset)
		return orders, total, err
	}

	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM orders"); err != nil {
		return nil, 0, err
	}
	err := r.db.SelectContext(ctx, &orders,
		"SELECT * FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2", perPage, offset)
	return orders, total, err
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id, status, note, createdBy, outboxTopic string, outboxPayload interface{}) error {
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

	if outboxTopic != "" {
		if err := outbox.InsertEvent(ctx, tx, outboxTopic, outboxPayload); err != nil {
			return err
		}
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

func (r *orderRepo) MarkCancelled(ctx context.Context, id, outboxTopic string, outboxPayload interface{}) error {
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

	if outboxTopic != "" {
		if err := outbox.InsertEvent(ctx, tx, outboxTopic, outboxPayload); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepo) GetDashboardStats(ctx context.Context) (totalOrdersToday int64, slaPercentage float64, err error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	// Get total orders today
	err = r.db.GetContext(ctx, &totalOrdersToday, "SELECT COUNT(*) FROM orders WHERE created_at >= $1 AND created_at < $2", today, tomorrow)
	if err != nil {
		return 0, 0, err
	}

	// Calculate SLA (Delivered vs Total, we'll use all time or just recent)
	var totalDelivered int64
	err = r.db.GetContext(ctx, &totalDelivered, "SELECT COUNT(*) FROM orders WHERE status = $1", events.OrderStatusDelivered)
	if err != nil {
		return 0, 0, err
	}
	
	var totalAll int64
	err = r.db.GetContext(ctx, &totalAll, "SELECT COUNT(*) FROM orders")
	if err != nil {
		return 0, 0, err
	}

	if totalAll > 0 {
		slaPercentage = float64(totalDelivered) / float64(totalAll) * 100.0
	} else {
		slaPercentage = 100.0 // Default to 100% if no orders yet
	}

	return totalOrdersToday, slaPercentage, nil
}

func (r *orderRepo) GetVolumeStats(ctx context.Context) ([]model.DailyVolume, error) {
	// Query for the last 7 days
	// Using PostgreSQL generate_series to ensure we get a row for every day even if 0
	query := `
		WITH dates AS (
			SELECT generate_series(
				current_date - interval '6 days',
				current_date,
				'1 day'::interval
			)::date AS date
		)
		SELECT 
			to_char(d.date, 'Mon DD') as date_str,
			COUNT(o.id) as count
		FROM dates d
		LEFT JOIN orders o ON DATE(o.created_at) = d.date
		GROUP BY d.date, date_str
		ORDER BY d.date ASC
	`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.DailyVolume
	for rows.Next() {
		var dv model.DailyVolume
		if err := rows.Scan(&dv.Date, &dv.Count); err != nil {
			return nil, err
		}
		result = append(result, dv)
	}
	return result, nil
}

