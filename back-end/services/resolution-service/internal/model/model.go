package model

import "time"

type Ticket struct {
	ID          string     `json:"id" db:"id"`
	OrderID     string     `json:"order_id" db:"order_id"`
	AWB         string     `json:"awb" db:"awb"`
	UserID      string     `json:"user_id" db:"user_id"`
	Type        string     `json:"type" db:"type"` // LOST, DAMAGED, DELIVERY_FAILED, COMPLAINT
	Priority    string     `json:"priority" db:"priority"` // LOW, MEDIUM, HIGH, CRITICAL
	Status      string     `json:"status" db:"status"` // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	Description string     `json:"description" db:"description"`
	Resolution  string     `json:"resolution,omitempty" db:"resolution"` // REFUND, RESEND, RETURN, CLOSED
	AgentID     *string    `json:"agent_id,omitempty" db:"agent_id"`
	Evidence    []string   `json:"evidence,omitempty" db:"-"` // photo URLs from MinIO
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

type Claim struct {
	ID           string    `json:"id" db:"id"`
	TicketID     string    `json:"ticket_id" db:"ticket_id"`
	OrderID      string    `json:"order_id" db:"order_id"`
	ClaimType    string    `json:"claim_type" db:"claim_type"` // INSURANCE, REFUND
	Amount       float64   `json:"amount" db:"amount"`
	Status       string    `json:"status" db:"status"` // PENDING, APPROVED, REJECTED, PAID
	ApprovedAt   *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type CreateTicketRequest struct {
	OrderID     string `json:"order_id"`
	AWB         string `json:"awb"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type UpdateTicketRequest struct {
	Status     string `json:"status,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	AgentID    string `json:"agent_id,omitempty"`
}

type CreateClaimRequest struct {
	TicketID  string  `json:"ticket_id"`
	OrderID   string  `json:"order_id"`
	ClaimType string  `json:"claim_type"`
	Amount    float64 `json:"amount"`
}

type UpdateClaimRequest struct {
	Status string  `json:"status"` // APPROVED, REJECTED, PAID
	Amount float64 `json:"amount"`
}
