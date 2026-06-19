package model

import "time"

type Courier struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	FullName      string    `json:"full_name" db:"full_name"`
	Phone         string    `json:"phone" db:"phone"`
	Email         string    `json:"email" db:"email"`
	VehicleType   string    `json:"vehicle_type" db:"vehicle_type"` // MOTORCYCLE, CAR, VAN
	VehiclePlate  string    `json:"vehicle_plate" db:"vehicle_plate"`
	MaxCapacityKg float64   `json:"max_capacity_kg" db:"max_capacity_kg"`
	CurrentLat    float64   `json:"current_lat" db:"current_lat"`
	CurrentLng    float64   `json:"current_lng" db:"current_lng"`
	IsOnline      bool      `json:"is_online" db:"is_online"`
	IsAvailable   bool      `json:"is_available" db:"is_available"` // not currently on a delivery
	Rating        float64   `json:"rating" db:"rating"`
	TotalDeliveries int    `json:"total_deliveries" db:"total_deliveries"`
	CoverageArea  string    `json:"coverage_area" db:"coverage_area"` // city/district
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterCourierRequest struct {
	UserID        string  `json:"user_id"`
	FullName      string  `json:"full_name"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	VehicleType   string  `json:"vehicle_type"`
	VehiclePlate  string  `json:"vehicle_plate"`
	MaxCapacityKg float64 `json:"max_capacity_kg"`
	CoverageArea  string  `json:"coverage_area"`
}

// EnsureCourierRequest get-or-creates a courier row for the authenticated user
// (UserID comes from the JWT/header, not the body) so a COURIER account is
// auto-linked to courier data on first use.
type EnsureCourierRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type UpdateStatusRequest struct {
	IsOnline bool `json:"is_online"`
}

type UpdateLocationRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
