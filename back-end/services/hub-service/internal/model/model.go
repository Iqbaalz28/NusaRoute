package model

import "time"

type Hub struct {
	ID       string  `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	Code     string  `json:"code" db:"code"`
	City     string  `json:"city" db:"city"`
	Province string  `json:"province" db:"province"`
	Lat      float64 `json:"lat" db:"lat"`
	Lng      float64 `json:"lng" db:"lng"`
	Type     string  `json:"type" db:"type"` // SORTATION, TRANSIT, DISTRIBUTION
	IsActive bool    `json:"is_active" db:"is_active"`
}

type ScanLog struct {
	ID         string    `json:"id" db:"id"`
	AWB        string    `json:"awb" db:"awb"`
	OrderID    string    `json:"order_id" db:"order_id"`
	HubID      string    `json:"hub_id" db:"hub_id"`
	ScanType   string    `json:"scan_type" db:"scan_type"` // ARRIVED, SORTED, DEPARTED
	OperatorID string    `json:"operator_id" db:"operator_id"`
	Note       string    `json:"note,omitempty" db:"note"`
	ScannedAt  time.Time `json:"scanned_at" db:"scanned_at"`
}

// HubUpsertRequest is the admin payload for creating/updating a hub. IsActive is a
// pointer so an update can leave it unchanged (nil) vs. explicitly set false.
type HubUpsertRequest struct {
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	City     string  `json:"city"`
	Province string  `json:"province"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Type     string  `json:"type"`
	IsActive *bool   `json:"is_active"`
}

type ScanRequest struct {
	AWB        string `json:"awb"`
	OrderID    string `json:"order_id"`
	HubID      string `json:"hub_id"`
	ScanType   string `json:"scan_type"`
	OperatorID string `json:"operator_id"`
	Note       string `json:"note,omitempty"`
}
