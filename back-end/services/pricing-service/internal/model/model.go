package model

// PriceCalculationRequest is the input for calculating shipping cost.
type PriceCalculationRequest struct {
	OriginLat    float64 `json:"origin_lat"`
	OriginLng    float64 `json:"origin_lng"`
	DestLat      float64 `json:"dest_lat"`
	DestLng      float64 `json:"dest_lng"`
	WeightKg     float64 `json:"weight_kg"`
	LengthCm     float64 `json:"length_cm"`
	WidthCm      float64 `json:"width_cm"`
	HeightCm     float64 `json:"height_cm"`
	ServiceType  string  `json:"service_type"` // REG, YES, CARGO, SAME
	IsInsured    bool    `json:"is_insured"`
	InsuredValue float64 `json:"insured_value"`
}

// PriceCalculationResponse is the output of pricing calculation.
type PriceCalculationResponse struct {
	ServiceType   string  `json:"service_type"`
	ServiceName   string  `json:"service_name"`
	DistanceKm    float64 `json:"distance_km"`
	WeightKg      float64 `json:"weight_kg"`
	VolumetricKg  float64 `json:"volumetric_kg"`
	ChargeableKg  float64 `json:"chargeable_kg"` // max(actual, volumetric)
	BaseCost      float64 `json:"base_cost"`
	WeightCost    float64 `json:"weight_cost"`
	InsuranceCost float64 `json:"insurance_cost"`
	Discount      float64 `json:"discount"`
	TotalCost     float64 `json:"total_cost"`
	EstDays       string  `json:"estimated_days"`
}

// ServiceInfo describes an available service type.
type ServiceInfo struct {
	Code         string  `json:"code" db:"code"`
	Name         string  `json:"name" db:"name"`
	Description  string  `json:"description" db:"description"`
	PricePerKm   float64 `json:"price_per_km" db:"price_per_km"`
	PricePerKg   float64 `json:"price_per_kg" db:"price_per_kg"`
	BaseFee      float64 `json:"base_fee" db:"base_fee"`
	EstDays      string  `json:"estimated_days" db:"est_days"`
	IsActive     bool    `json:"is_active" db:"is_active"`
}

// TariffZone represents a zone-based pricing rule.
type TariffZone struct {
	ID           string  `json:"id" db:"id"`
	ZoneName     string  `json:"zone_name" db:"zone_name"`
	MinDistKm    float64 `json:"min_dist_km" db:"min_dist_km"`
	MaxDistKm    float64 `json:"max_dist_km" db:"max_dist_km"`
	Multiplier   float64 `json:"multiplier" db:"multiplier"`
}
