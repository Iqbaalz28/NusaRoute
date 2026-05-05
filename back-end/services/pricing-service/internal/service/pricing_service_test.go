//go:build unit

package service_test

import (
	"context"
	"testing"
	"math"

	"github.com/nusaroute/services/pricing-service/internal/model"
	"github.com/nusaroute/services/pricing-service/internal/service"
)

// MockPricingRepository for unit testing
type MockPricingRepository struct{}

func (m *MockPricingRepository) GetServiceByCode(ctx context.Context, code string) (*model.ServiceInfo, error) {
	services := map[string]*model.ServiceInfo{
		"REG":  {Code: "REG", Name: "Reguler", PricePerKm: 30, PricePerKg: 2500, BaseFee: 8000, EstDays: "2-4 hari", IsActive: true},
		"YES":  {Code: "YES", Name: "Yakin Esok Sampai", PricePerKm: 80, PricePerKg: 5000, BaseFee: 15000, EstDays: "1 hari", IsActive: true},
		"CARGO": {Code: "CARGO", Name: "Kargo", PricePerKm: 15, PricePerKg: 1500, BaseFee: 5000, EstDays: "5-7 hari", IsActive: true},
		"SAME": {Code: "SAME", Name: "Same Day", PricePerKm: 150, PricePerKg: 8000, BaseFee: 25000, EstDays: "< 12 jam", IsActive: true},
	}
	if svc, ok := services[code]; ok { return svc, nil }
	return nil, nil
}

func (m *MockPricingRepository) GetAllServices(ctx context.Context) ([]model.ServiceInfo, error) {
	return []model.ServiceInfo{
		{Code: "REG", Name: "Reguler", PricePerKm: 30, PricePerKg: 2500, BaseFee: 8000, EstDays: "2-4 hari", IsActive: true},
		{Code: "YES", Name: "Yakin Esok Sampai", PricePerKm: 80, PricePerKg: 5000, BaseFee: 15000, EstDays: "1 hari", IsActive: true},
	}, nil
}

func (m *MockPricingRepository) GetTariffZone(ctx context.Context, distanceKm float64) (*model.TariffZone, error) {
	if distanceKm <= 30 { return &model.TariffZone{Multiplier: 1.0}, nil }
	if distanceKm <= 150 { return &model.TariffZone{Multiplier: 1.2}, nil }
	return &model.TariffZone{Multiplier: 1.5}, nil
}

func TestCalculate_Reguler_IntraCity(t *testing.T) {
	repo := &MockPricingRepository{}
	svc := service.NewPricingService(repo, nil) // nil Redis = no cache

	// Bandung to Bandung (~10km)
	result, err := svc.Calculate(context.Background(), model.PriceCalculationRequest{
		OriginLat: -6.917, OriginLng: 107.619, DestLat: -6.905, DestLng: 107.635,
		WeightKg: 2.0, LengthCm: 30, WidthCm: 20, HeightCm: 15,
		ServiceType: "REG",
	})

	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if result.ServiceType != "REG" { t.Errorf("wrong service type: %s", result.ServiceType) }
	if result.TotalCost <= 0 { t.Error("total cost should be positive") }
	if result.DistanceKm > 30 { t.Errorf("distance too large for intra-city: %.2f", result.DistanceKm) }
	t.Logf("Bandung intra-city: %.2f km, Rp %.0f", result.DistanceKm, result.TotalCost)
}

func TestCalculate_VolumetricWeight(t *testing.T) {
	repo := &MockPricingRepository{}
	svc := service.NewPricingService(repo, nil)

	// Large but light package: 60x40x30cm = 12kg volumetric
	result, err := svc.Calculate(context.Background(), model.PriceCalculationRequest{
		OriginLat: -6.917, OriginLng: 107.619, DestLat: -6.905, DestLng: 107.635,
		WeightKg: 1.0, LengthCm: 60, WidthCm: 40, HeightCm: 30,
		ServiceType: "REG",
	})

	if err != nil { t.Fatalf("unexpected error: %v", err) }
	expectedVolumetric := (60.0 * 40.0 * 30.0) / 6000.0 // = 12kg
	if math.Abs(result.VolumetricKg - expectedVolumetric) > 0.1 {
		t.Errorf("volumetric weight: expected %.2f, got %.2f", expectedVolumetric, result.VolumetricKg)
	}
	if result.ChargeableKg < result.VolumetricKg {
		t.Error("chargeable weight should be >= volumetric weight")
	}
}

func TestCalculate_WithInsurance(t *testing.T) {
	repo := &MockPricingRepository{}
	svc := service.NewPricingService(repo, nil)

	result, err := svc.Calculate(context.Background(), model.PriceCalculationRequest{
		OriginLat: -6.917, OriginLng: 107.619, DestLat: -7.257, DestLng: 112.752,
		WeightKg: 3.0, ServiceType: "REG",
		IsInsured: true, InsuredValue: 5000000, // 5 juta
	})

	if err != nil { t.Fatalf("unexpected error: %v", err) }
	expectedInsurance := 5000000.0 * 0.002 // 0.2% = 10.000
	if math.Abs(result.InsuranceCost - expectedInsurance) > 100 {
		t.Errorf("insurance: expected ~%.0f, got %.0f", expectedInsurance, result.InsuranceCost)
	}
}

func TestCalculate_InvalidWeight(t *testing.T) {
	repo := &MockPricingRepository{}
	svc := service.NewPricingService(repo, nil)

	_, err := svc.Calculate(context.Background(), model.PriceCalculationRequest{
		WeightKg: 0, ServiceType: "REG",
	})
	if err == nil { t.Error("expected error for zero weight") }
}

func TestCalculate_MissingServiceType(t *testing.T) {
	repo := &MockPricingRepository{}
	svc := service.NewPricingService(repo, nil)

	_, err := svc.Calculate(context.Background(), model.PriceCalculationRequest{
		WeightKg: 1.0, ServiceType: "",
	})
	if err == nil { t.Error("expected error for empty service type") }
}

func TestCompareAll(t *testing.T) {
	repo := &MockPricingRepository{}
	svc := service.NewPricingService(repo, nil)

	results, err := svc.CalculateAll(context.Background(), model.PriceCalculationRequest{
		OriginLat: -6.917, OriginLng: 107.619, DestLat: -6.905, DestLng: 107.635,
		WeightKg: 2.0,
	})

	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(results) < 2 { t.Errorf("expected at least 2 results, got %d", len(results)) }
}
