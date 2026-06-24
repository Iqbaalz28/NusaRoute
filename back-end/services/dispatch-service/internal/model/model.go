package model

import "time"

type Assignment struct {
	ID          string     `json:"id" db:"id"`
	OrderID     string     `json:"order_id" db:"order_id"`
	AWB         string     `json:"awb" db:"awb"`
	CourierID   string     `json:"courier_id" db:"courier_id"`
	CourierName string     `json:"courier_name" db:"courier_name"`
	Status      string     `json:"status" db:"status"` // OPEN, ASSIGNED, PICKED_UP, COMPLETED, NO_SHOW, REASSIGNED
	Leg         string     `json:"leg" db:"leg"`       // FIRST_MILE, LAST_MILE, DIRECT
	PickupLat   float64    `json:"pickup_lat" db:"pickup_lat"`
	PickupLng   float64    `json:"pickup_lng" db:"pickup_lng"`
	PickupAddr  string     `json:"pickup_address" db:"pickup_address"`
	DropoffLat  float64    `json:"dropoff_lat" db:"dropoff_lat"`
	DropoffLng  float64    `json:"dropoff_lng" db:"dropoff_lng"`
	DropoffAddr string     `json:"dropoff_address" db:"dropoff_address"`
	HubID       string     `json:"hub_id" db:"hub_id"`     // the hub involved in this leg (origin for first-mile, dest for last-mile)
	HubName     string     `json:"hub_name" db:"hub_name"`
	AssignedAt  time.Time  `json:"assigned_at" db:"assigned_at"`
	PickedUpAt  *time.Time `json:"picked_up_at,omitempty" db:"picked_up_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

const (
	AssignmentStatusOpen       = "OPEN" // job available, no courier yet (take-job)
	AssignmentStatusAssigned   = "ASSIGNED"
	AssignmentStatusPickedUp   = "PICKED_UP"
	AssignmentStatusCompleted  = "COMPLETED"
	AssignmentStatusNoShow     = "NO_SHOW"
	AssignmentStatusReassigned = "REASSIGNED"

	// Leg types for the hub-and-spoke flow.
	LegFirstMile = "FIRST_MILE" // sender → origin hub
	LegLastMile  = "LAST_MILE"  // dest hub → receiver
	LegDirect    = "DIRECT"     // sender → receiver (same-city instant, no hub)
)

type ManualAssignRequest struct {
	OrderID   string `json:"order_id"`
	CourierID string `json:"courier_id"`
}

// ClaimRequest is sent by a courier to claim an OPEN job (first-come-first-served).
// CourierID is taken from the authenticated identity, not the body.
type ClaimRequest struct {
	OrderID     string `json:"order_id"`
	CourierName string `json:"courier_name"` // display name from the courier's profile
}

// PickupRequest is sent by a courier (driver) when they pick up a package from the
// sender. AWB is the code scanned from the sender's digital label and must match
// the claimed assignment.
type PickupRequest struct {
	OrderID string  `json:"order_id"`
	AWB     string  `json:"awb"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// DeliverRequest is sent by a courier when completing delivery to the receiver.
// AWB is the code scanned from the package label and must match the assignment.
type DeliverRequest struct {
	OrderID      string  `json:"order_id"`
	AWB          string  `json:"awb"`
	ReceiverName string  `json:"receiver_name"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
}
