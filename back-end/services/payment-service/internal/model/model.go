package model

import "time"

// Transaction represents a payment transaction record.
type Transaction struct {
	ID              string    `json:"id" db:"id"`
	OrderID         string    `json:"order_id" db:"order_id"`
	Amount          float64   `json:"amount" db:"amount"`
	Method          string    `json:"method" db:"method"` // VA, E_WALLET, CARD, COD
	Status          string    `json:"status" db:"status"` // PENDING, PAID, FAILED, REFUNDED
	PaymentURL      string    `json:"payment_url,omitempty" db:"payment_url"`
	ExternalID      string    `json:"external_id,omitempty" db:"external_id"` // ID from payment gateway
	IdempotencyKey  string    `json:"idempotency_key" db:"idempotency_key"`
	PaidAt          *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	FailedAt        *time.Time `json:"failed_at,omitempty" db:"failed_at"`
	RefundedAt      *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// InitiatePaymentRequest is the input for starting a payment.
type InitiatePaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Method  string  `json:"method"`
}

// WebhookPayload simulates the payload from a payment gateway webhook.
type WebhookPayload struct {
	ExternalID     string  `json:"external_id"`
	OrderID        string  `json:"order_id"`
	Status         string  `json:"status"` // PAID, FAILED
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotency_key"`
}

// Payment statuses
const (
	PaymentStatusPending  = "PENDING"
	PaymentStatusPaid     = "PAID"
	PaymentStatusFailed   = "FAILED"
	PaymentStatusRefunded = "REFUNDED"
)
