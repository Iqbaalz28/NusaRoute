package model

import "time"

type Assignment struct {
	ID          string     `json:"id" db:"id"`
	OrderID     string     `json:"order_id" db:"order_id"`
	AWB         string     `json:"awb" db:"awb"`
	CourierID   string     `json:"courier_id" db:"courier_id"`
	CourierName string     `json:"courier_name" db:"courier_name"`
	Status      string     `json:"status" db:"status"` // ASSIGNED, PICKED_UP, COMPLETED, NO_SHOW, REASSIGNED
	PickupLat   float64    `json:"pickup_lat" db:"pickup_lat"`
	PickupLng   float64    `json:"pickup_lng" db:"pickup_lng"`
	PickupAddr  string     `json:"pickup_address" db:"pickup_address"`
	AssignedAt  time.Time  `json:"assigned_at" db:"assigned_at"`
	PickedUpAt  *time.Time `json:"picked_up_at,omitempty" db:"picked_up_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

const (
	AssignmentStatusAssigned   = "ASSIGNED"
	AssignmentStatusPickedUp   = "PICKED_UP"
	AssignmentStatusCompleted  = "COMPLETED"
	AssignmentStatusNoShow     = "NO_SHOW"
	AssignmentStatusReassigned = "REASSIGNED"
)

type ManualAssignRequest struct {
	OrderID   string `json:"order_id"`
	CourierID string `json:"courier_id"`
}

// PickupRequest is sent by a courier (driver) when they pick up a package from the sender.
type PickupRequest struct {
	OrderID string  `json:"order_id"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// DeliverRequest is sent by a courier (driver) when they complete delivery to the receiver.
type DeliverRequest struct {
	OrderID      string  `json:"order_id"`
	ReceiverName string  `json:"receiver_name"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
}
