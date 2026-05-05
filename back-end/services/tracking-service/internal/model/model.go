package model

import "time"

// TrackingEvent is an immutable record in the tracking timeline (MongoDB document).
type TrackingEvent struct {
	ID        string    `json:"id" bson:"_id"`
	AWB       string    `json:"awb" bson:"awb"`
	OrderID   string    `json:"order_id" bson:"order_id"`
	Status    string    `json:"status" bson:"status"`
	Location  string    `json:"location" bson:"location"`
	Detail    string    `json:"detail" bson:"detail"`
	Lat       float64   `json:"lat,omitempty" bson:"lat,omitempty"`
	Lng       float64   `json:"lng,omitempty" bson:"lng,omitempty"`
	Source    string    `json:"source" bson:"source"` // which service produced this event
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// CourierGPS represents a live GPS ping from a courier (stored in Redis).
type CourierGPS struct {
	CourierID string  `json:"courier_id"`
	AWB       string  `json:"awb"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Timestamp int64   `json:"timestamp"`
}

// TrackingTimeline is the response for a tracking query.
type TrackingTimeline struct {
	AWB       string          `json:"awb"`
	OrderID   string          `json:"order_id"`
	Status    string          `json:"current_status"`
	Events    []TrackingEvent `json:"events"`
	LiveGPS   *CourierGPS     `json:"live_gps,omitempty"`
}

type GPSUpdateRequest struct {
	CourierID string  `json:"courier_id"`
	AWB       string  `json:"awb"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}
