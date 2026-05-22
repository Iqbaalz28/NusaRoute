package service

import (
	"github.com/nusaroute/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/nusaroute/services/pricing-service/internal/model"
	"github.com/nusaroute/services/pricing-service/internal/repository"
)

const (
	insuranceRate      = 0.002  // 0.2% of insured value
	volumetricDivisor  = 6000.0 // industry standard
	cacheTTL           = 30 * time.Minute
)

type PricingService interface {
	Calculate(ctx context.Context, req model.PriceCalculationRequest) (*model.PriceCalculationResponse, error)
	CalculateAll(ctx context.Context, req model.PriceCalculationRequest) ([]model.PriceCalculationResponse, error)
	GetServices(ctx context.Context) ([]model.ServiceInfo, error)
}

type pricingService struct {
	repo  repository.PricingRepository
	cache *redis.Client
}

func NewPricingService(repo repository.PricingRepository, cache *redis.Client) PricingService {
	return &pricingService{repo: repo, cache: cache}
}

func (s *pricingService) Calculate(ctx context.Context, req model.PriceCalculationRequest) (*model.PriceCalculationResponse, error) {
	if req.WeightKg <= 0 {
		return nil, errors.New("weight must be positive")
	}
	if req.ServiceType == "" {
		return nil, errors.New("service_type is required")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("price:%s:%.4f:%.4f:%.4f:%.4f:%.2f",
		req.ServiceType, req.OriginLat, req.OriginLng, req.DestLat, req.DestLng, req.WeightKg)

	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp model.PriceCalculationResponse
			if json.Unmarshal([]byte(cached), &resp) == nil {
				logger.Info(context.Background(), fmt.Sprintf("[Pricing] Cache HIT for %s", cacheKey))
				return &resp, nil
			}
		}
	}

	// Get service config
	svcInfo, err := s.repo.GetServiceByCode(ctx, req.ServiceType)
	if err != nil {
		return nil, fmt.Errorf("service type %s not found", req.ServiceType)
	}

	// Calculate distance using Haversine formula
	distanceKm := haversine(req.OriginLat, req.OriginLng, req.DestLat, req.DestLng)

	// Get tariff zone multiplier
	zone, _ := s.repo.GetTariffZone(ctx, distanceKm)

	// Calculate volumetric weight
	volumetricKg := (req.LengthCm * req.WidthCm * req.HeightCm) / volumetricDivisor
	chargeableKg := math.Max(req.WeightKg, volumetricKg)

	// Calculate costs
	baseCost := svcInfo.BaseFee
	weightCost := chargeableKg * svcInfo.PricePerKg * zone.Multiplier
	distanceCost := distanceKm * svcInfo.PricePerKm

	var insuranceCost float64
	if req.IsInsured && req.InsuredValue > 0 {
		insuranceCost = req.InsuredValue * insuranceRate
	}

	totalCost := baseCost + weightCost + distanceCost + insuranceCost

	// Round to nearest 500
	totalCost = math.Ceil(totalCost/500) * 500

	resp := &model.PriceCalculationResponse{
		ServiceType:  svcInfo.Code,
		ServiceName:  svcInfo.Name,
		DistanceKm:   math.Round(distanceKm*100) / 100,
		WeightKg:     req.WeightKg,
		VolumetricKg: math.Round(volumetricKg*100) / 100,
		ChargeableKg: math.Round(chargeableKg*100) / 100,
		BaseCost:     baseCost,
		WeightCost:   math.Round((weightCost+distanceCost)*100) / 100,
		InsuranceCost: math.Round(insuranceCost*100) / 100,
		Discount:     0,
		TotalCost:    totalCost,
		EstDays:      svcInfo.EstDays,
	}

	// Cache result
	if s.cache != nil {
		data, _ := json.Marshal(resp)
		s.cache.Set(ctx, cacheKey, string(data), cacheTTL)
	}

	return resp, nil
}

func (s *pricingService) CalculateAll(ctx context.Context, req model.PriceCalculationRequest) ([]model.PriceCalculationResponse, error) {
	services, err := s.repo.GetAllServices(ctx)
	if err != nil {
		return nil, err
	}

	var results []model.PriceCalculationResponse
	for _, svc := range services {
		req.ServiceType = svc.Code
		result, err := s.Calculate(ctx, req)
		if err != nil {
			continue
		}
		results = append(results, *result)
	}
	return results, nil
}

func (s *pricingService) GetServices(ctx context.Context) ([]model.ServiceInfo, error) {
	return s.repo.GetAllServices(ctx)
}

// haversine calculates the distance in km between two lat/lng points.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
