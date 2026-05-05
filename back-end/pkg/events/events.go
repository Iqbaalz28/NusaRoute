// Package events defines all Kafka event types and schemas used across NusaRoute microservices.
// This is the single source of truth for event contracts between services.
package events

import "time"

// ============================================================
// Kafka Topic Names
// ============================================================

const (
	// Payment events
	TopicPaymentSuccess = "payment.success"
	TopicPaymentFailed  = "payment.failed"

	// Order events
	TopicOrderReadyForPickup = "order.ready-for-pickup"
	TopicOrderCancelled      = "order.cancelled"

	// Courier/Dispatch events
	TopicCourierAssigned   = "courier.assigned"
	TopicCourierReassigned = "courier.reassigned"

	// Package lifecycle events
	TopicPackagePickedUp    = "package.picked-up"
	TopicPackageScannedHub  = "package.scanned-at-hub"
	TopicPackageDelivered   = "package.delivered"
	TopicDeliveryFailed     = "delivery.failed"
	TopicPackageLost        = "package.lost.suspected"
	TopicPackageDamaged     = "package.damaged"

	// Resolution events
	TopicResolutionCreated  = "resolution.created"
	TopicResolutionResolved = "resolution.resolved"
	TopicOrderReturned      = "order.returned"

	// Dead Letter Queue
	TopicDLQ = "nusaroute.dlq"
)

// ============================================================
// Event Schemas
// ============================================================

// BaseEvent contains common fields for all events.
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // service name that produced the event
}

// PaymentSuccessEvent is published when payment is confirmed.
type PaymentSuccessEvent struct {
	BaseEvent
	OrderID       string  `json:"order_id"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Method        string  `json:"method"` // VA, E_WALLET, CARD
}

// PaymentFailedEvent is published when payment fails.
type PaymentFailedEvent struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// OrderReadyForPickupEvent is published when an order is ready for courier pickup.
type OrderReadyForPickupEvent struct {
	BaseEvent
	OrderID      string  `json:"order_id"`
	AWB          string  `json:"awb"`
	SenderID     string  `json:"sender_id"`
	PickupLat    float64 `json:"pickup_lat"`
	PickupLng    float64 `json:"pickup_lng"`
	PickupAddr   string  `json:"pickup_address"`
	ReceiverAddr string  `json:"receiver_address"`
	ReceiverLat  float64 `json:"receiver_lat"`
	ReceiverLng  float64 `json:"receiver_lng"`
	Weight       float64 `json:"weight_kg"`
	ServiceType  string  `json:"service_type"` // REG, YES, CARGO
}

// OrderCancelledEvent is published when an order is cancelled.
type OrderCancelledEvent struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// CourierAssignedEvent is published when a courier is assigned to a pickup.
type CourierAssignedEvent struct {
	BaseEvent
	OrderID      string `json:"order_id"`
	AWB          string `json:"awb"`
	CourierID    string `json:"courier_id"`
	CourierName  string `json:"courier_name"`
	CourierPhone string `json:"courier_phone"`
	EstimatedPickupTime time.Time `json:"estimated_pickup_time"`
}

// CourierReassignedEvent is published when a courier is reassigned (no-show).
type CourierReassignedEvent struct {
	BaseEvent
	OrderID        string `json:"order_id"`
	OldCourierID   string `json:"old_courier_id"`
	NewCourierID   string `json:"new_courier_id"`
	NewCourierName string `json:"new_courier_name"`
	Reason         string `json:"reason"`
}

// PackagePickedUpEvent is published when courier scans and picks up a package.
type PackagePickedUpEvent struct {
	BaseEvent
	OrderID   string    `json:"order_id"`
	AWB       string    `json:"awb"`
	CourierID string    `json:"courier_id"`
	PickedAt  time.Time `json:"picked_at"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
}

// PackageScannedAtHubEvent is published for every hub scan (inbound/sort/outbound).
type PackageScannedAtHubEvent struct {
	BaseEvent
	OrderID   string `json:"order_id"`
	AWB       string `json:"awb"`
	HubID     string `json:"hub_id"`
	HubName   string `json:"hub_name"`
	ScanType  string `json:"scan_type"` // ARRIVED, SORTED, DEPARTED
	OperatorID string `json:"operator_id"`
}

// PackageDeliveredEvent is published when courier completes delivery.
type PackageDeliveredEvent struct {
	BaseEvent
	OrderID      string  `json:"order_id"`
	AWB          string  `json:"awb"`
	CourierID    string  `json:"courier_id"`
	ReceiverName string  `json:"receiver_name"`
	ProofPhotoURL string `json:"proof_photo_url"` // Proof of Delivery
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	GeofenceValid bool   `json:"geofence_valid"`
}

// DeliveryFailedEvent is published when delivery attempt fails.
type DeliveryFailedEvent struct {
	BaseEvent
	OrderID     string `json:"order_id"`
	AWB         string `json:"awb"`
	CourierID   string `json:"courier_id"`
	Reason      string `json:"reason"` // ADDRESS_NOT_FOUND, RECIPIENT_ABSENT, REFUSED
	AttemptNum  int    `json:"attempt_number"`
	MaxAttempts int    `json:"max_attempts"`
}

// PackageLostSuspectedEvent is published when a package exceeds SLA without updates.
type PackageLostSuspectedEvent struct {
	BaseEvent
	OrderID          string    `json:"order_id"`
	AWB              string    `json:"awb"`
	LastKnownHubID   string    `json:"last_known_hub_id"`
	LastKnownHubName string    `json:"last_known_hub_name"`
	LastScanTime     time.Time `json:"last_scan_time"`
	HoursSinceUpdate int       `json:"hours_since_update"`
}

// PackageDamagedEvent is published when damage is reported.
type PackageDamagedEvent struct {
	BaseEvent
	OrderID        string `json:"order_id"`
	AWB            string `json:"awb"`
	ReportedBy     string `json:"reported_by"` // COURIER or HUB_OPERATOR
	ReporterID     string `json:"reporter_id"`
	DamageDesc     string `json:"damage_description"`
	EvidencePhotos []string `json:"evidence_photos"`
}

// ResolutionCreatedEvent is published when a resolution ticket is opened.
type ResolutionCreatedEvent struct {
	BaseEvent
	TicketID  string `json:"ticket_id"`
	OrderID   string `json:"order_id"`
	AWB       string `json:"awb"`
	Type      string `json:"type"` // LOST, DAMAGED, DELIVERY_FAILED, COMPLAINT
	Priority  string `json:"priority"` // LOW, MEDIUM, HIGH, CRITICAL
}

// ResolutionResolvedEvent is published when a resolution is completed.
type ResolutionResolvedEvent struct {
	BaseEvent
	TicketID   string `json:"ticket_id"`
	OrderID    string `json:"order_id"`
	Resolution string `json:"resolution"` // REFUND, RESEND, RETURN, CLOSED
}

// OrderReturnedEvent is published when a return is initiated.
type OrderReturnedEvent struct {
	BaseEvent
	OrderID    string `json:"order_id"`
	AWB        string `json:"awb"`
	ReturnAWB  string `json:"return_awb"`
	Reason     string `json:"reason"`
}

// ============================================================
// Order Status Constants
// ============================================================

const (
	OrderStatusDraft            = "DRAFT"
	OrderStatusPendingPayment   = "PENDING_PAYMENT"
	OrderStatusReadyForPickup   = "READY_FOR_PICKUP"
	OrderStatusPickedUp         = "PICKED_UP"
	OrderStatusInTransit        = "IN_TRANSIT"
	OrderStatusOutForDelivery   = "OUT_FOR_DELIVERY"
	OrderStatusDelivered        = "DELIVERED"
	OrderStatusDeliveryFailed   = "DELIVERY_FAILED"
	OrderStatusReturnToSender   = "RETURN_TO_SENDER"
	OrderStatusCancelled        = "CANCELLED"
	OrderStatusLostSuspected    = "LOST_SUSPECTED"
)

// Hub scan types
const (
	ScanTypeArrived  = "ARRIVED"
	ScanTypeSorted   = "SORTED"
	ScanTypeDeparted = "DEPARTED"
)

// Delivery failure reasons
const (
	FailReasonAddressNotFound  = "ADDRESS_NOT_FOUND"
	FailReasonRecipientAbsent  = "RECIPIENT_ABSENT"
	FailReasonRecipientRefused = "REFUSED"
	FailReasonAccessBlocked    = "ACCESS_BLOCKED"

	MaxDeliveryAttempts = 3
)

// Service types
const (
	ServiceTypeREG   = "REG"   // Reguler (2-4 hari)
	ServiceTypeYES   = "YES"   // Yakin Esok Sampai (1 hari)
	ServiceTypeCARGO = "CARGO" // Kargo (5-7 hari)
	ServiceTypeSAME  = "SAME"  // Same Day
)

// Resolution types
const (
	ResolutionTypeLost           = "LOST"
	ResolutionTypeDamaged        = "DAMAGED"
	ResolutionTypeDeliveryFailed = "DELIVERY_FAILED"
	ResolutionTypeComplaint      = "COMPLAINT"
)

// Resolution statuses
const (
	ResolutionStatusOpen       = "OPEN"
	ResolutionStatusInProgress = "IN_PROGRESS"
	ResolutionStatusResolved   = "RESOLVED"
	ResolutionStatusClosed     = "CLOSED"
)
