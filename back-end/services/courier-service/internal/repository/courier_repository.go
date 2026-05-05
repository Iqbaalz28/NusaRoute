package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/services/courier-service/internal/model"
)

type CourierRepository interface {
	Create(ctx context.Context, courier *model.Courier) error
	GetByID(ctx context.Context, id string) (*model.Courier, error)
	GetByUserID(ctx context.Context, userID string) (*model.Courier, error)
	GetAvailable(ctx context.Context, lat, lng float64, radiusKm float64) ([]model.Courier, error)
	UpdateStatus(ctx context.Context, id string, isOnline bool) error
	UpdateLocation(ctx context.Context, id string, lat, lng float64) error
	SetAvailability(ctx context.Context, id string, available bool) error
	IncrementDeliveries(ctx context.Context, id string) error
}

type courierRepo struct{ db *sqlx.DB }

func NewCourierRepository(db *sqlx.DB) CourierRepository { return &courierRepo{db: db} }

func (r *courierRepo) Create(ctx context.Context, c *model.Courier) error {
	c.ID = uuid.New().String()
	c.IsActive = true
	c.IsOnline = false
	c.IsAvailable = true
	c.Rating = 5.0
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	query := `INSERT INTO couriers (id, user_id, full_name, phone, email, vehicle_type, vehicle_plate,
		max_capacity_kg, current_lat, current_lng, is_online, is_available, rating, total_deliveries,
		coverage_area, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.UserID, c.FullName, c.Phone, c.Email, c.VehicleType, c.VehiclePlate,
		c.MaxCapacityKg, c.CurrentLat, c.CurrentLng, c.IsOnline, c.IsAvailable, c.Rating,
		c.TotalDeliveries, c.CoverageArea, c.IsActive, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *courierRepo) GetByID(ctx context.Context, id string) (*model.Courier, error) {
	var c model.Courier
	err := r.db.GetContext(ctx, &c, "SELECT * FROM couriers WHERE id = $1 AND is_active = true", id)
	return &c, err
}

func (r *courierRepo) GetByUserID(ctx context.Context, userID string) (*model.Courier, error) {
	var c model.Courier
	err := r.db.GetContext(ctx, &c, "SELECT * FROM couriers WHERE user_id = $1 AND is_active = true", userID)
	return &c, err
}

func (r *courierRepo) GetAvailable(ctx context.Context, lat, lng, radiusKm float64) ([]model.Courier, error) {
	// Simple distance filter using Haversine approximation in SQL
	var couriers []model.Courier
	query := `SELECT *, 
		(6371 * acos(cos(radians($1)) * cos(radians(current_lat)) * cos(radians(current_lng) - radians($2)) 
		+ sin(radians($1)) * sin(radians(current_lat)))) AS distance_km
		FROM couriers
		WHERE is_online = true AND is_available = true AND is_active = true
		HAVING distance_km <= $3
		ORDER BY distance_km ASC
		LIMIT 20`
	
	// Fallback for PostgreSQL (no HAVING on alias)
	query = `SELECT * FROM couriers 
		WHERE is_online = true AND is_available = true AND is_active = true
		AND (6371 * acos(LEAST(1.0, cos(radians($1)) * cos(radians(current_lat)) * cos(radians(current_lng) - radians($2)) 
		+ sin(radians($1)) * sin(radians(current_lat))))) <= $3
		ORDER BY (6371 * acos(LEAST(1.0, cos(radians($1)) * cos(radians(current_lat)) * cos(radians(current_lng) - radians($2)) 
		+ sin(radians($1)) * sin(radians(current_lat))))) ASC
		LIMIT 20`

	err := r.db.SelectContext(ctx, &couriers, query, lat, lng, radiusKm)
	return couriers, err
}

func (r *courierRepo) UpdateStatus(ctx context.Context, id string, isOnline bool) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE couriers SET is_online = $1, updated_at = $2 WHERE id = $3", isOnline, time.Now(), id)
	return err
}

func (r *courierRepo) UpdateLocation(ctx context.Context, id string, lat, lng float64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE couriers SET current_lat = $1, current_lng = $2, updated_at = $3 WHERE id = $4",
		lat, lng, time.Now(), id)
	return err
}

func (r *courierRepo) SetAvailability(ctx context.Context, id string, available bool) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE couriers SET is_available = $1, updated_at = $2 WHERE id = $3", available, time.Now(), id)
	return err
}

func (r *courierRepo) IncrementDeliveries(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE couriers SET total_deliveries = total_deliveries + 1, updated_at = $1 WHERE id = $2",
		time.Now(), id)
	return err
}
