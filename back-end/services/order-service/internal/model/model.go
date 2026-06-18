package model

import "time"

// Order represents a shipment order.
type Order struct {
	ID             string    `json:"id" db:"id"`
	AWB            string    `json:"awb" db:"awb"` // Airway Bill number
	UserID         string    `json:"user_id" db:"user_id"`
	Status         string    `json:"status" db:"status"`
	ServiceType    string    `json:"service_type" db:"service_type"` // REG, YES, CARGO, SAME
	DeliveryMode   string    `json:"delivery_mode" db:"delivery_mode"` // DIRECT (same-city instant) or VIA_HUB

	// Sender info
	SenderName     string    `json:"sender_name" db:"sender_name"`
	SenderPhone    string    `json:"sender_phone" db:"sender_phone"`
	SenderAddress  string    `json:"sender_address" db:"sender_address"`
	SenderLat      float64   `json:"sender_lat" db:"sender_lat"`
	SenderLng      float64   `json:"sender_lng" db:"sender_lng"`
	
	// Receiver info
	ReceiverName   string    `json:"receiver_name" db:"receiver_name"`
	ReceiverPhone  string    `json:"receiver_phone" db:"receiver_phone"`
	ReceiverAddress string   `json:"receiver_address" db:"receiver_address"`
	ReceiverLat    float64   `json:"receiver_lat" db:"receiver_lat"`
	ReceiverLng    float64   `json:"receiver_lng" db:"receiver_lng"`
	
	// Package details
	ItemDescription string   `json:"item_description" db:"item_description"`
	WeightKg       float64   `json:"weight_kg" db:"weight_kg"`
	LengthCm       float64   `json:"length_cm" db:"length_cm"`
	WidthCm        float64   `json:"width_cm" db:"width_cm"`
	HeightCm       float64   `json:"height_cm" db:"height_cm"`
	IsInsured      bool      `json:"is_insured" db:"is_insured"`
	InsuredValue   float64   `json:"insured_value" db:"insured_value"`
	
	// Pricing
	ShippingCost   float64   `json:"shipping_cost" db:"shipping_cost"`
	InsuranceCost  float64   `json:"insurance_cost" db:"insurance_cost"`
	TotalCost      float64   `json:"total_cost" db:"total_cost"`
	
	// Delivery tracking
	DeliveryAttempts int     `json:"delivery_attempts" db:"delivery_attempts"`
	CourierID      *string   `json:"courier_id,omitempty" db:"courier_id"`
	
	// Timestamps
	PaidAt         *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	PickedUpAt     *time.Time `json:"picked_up_at,omitempty" db:"picked_up_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// OrderStatusHistory records each status change for audit.
type OrderStatusHistory struct {
	ID        string    `json:"id" db:"id"`
	OrderID   string    `json:"order_id" db:"order_id"`
	Status    string    `json:"status" db:"status"`
	Note      string    `json:"note" db:"note"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateOrderRequest is the input for creating a new order.
type CreateOrderRequest struct {
	// Sender
	SenderName    string  `json:"sender_name"`
	SenderPhone   string  `json:"sender_phone"`
	SenderAddress string  `json:"sender_address"`
	SenderLat     float64 `json:"sender_lat"`
	SenderLng     float64 `json:"sender_lng"`
	
	// Receiver
	ReceiverName    string  `json:"receiver_name"`
	ReceiverPhone   string  `json:"receiver_phone"`
	ReceiverAddress string  `json:"receiver_address"`
	ReceiverLat     float64 `json:"receiver_lat"`
	ReceiverLng     float64 `json:"receiver_lng"`
	
	// Package
	ItemDescription string  `json:"item_description"`
	WeightKg        float64 `json:"weight_kg"`
	LengthCm        float64 `json:"length_cm"`
	WidthCm         float64 `json:"width_cm"`
	HeightCm        float64 `json:"height_cm"`
	ServiceType     string  `json:"service_type"`
	IsInsured       bool    `json:"is_insured"`
	InsuredValue    float64 `json:"insured_value"`
	
	// Pricing (pre-calculated by Pricing Service)
	ShippingCost  float64 `json:"shipping_cost"`
	InsuranceCost float64 `json:"insurance_cost"`
	TotalCost     float64 `json:"total_cost"`
}

// DailyVolume represents daily order statistics
type DailyVolume struct {
	Date  string `json:"date" db:"date"`
	Count int64  `json:"count" db:"count"`
}
