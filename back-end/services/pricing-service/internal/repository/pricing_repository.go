package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/services/pricing-service/internal/model"
)

type PricingRepository interface {
	GetServiceByCode(ctx context.Context, code string) (*model.ServiceInfo, error)
	GetAllServices(ctx context.Context) ([]model.ServiceInfo, error)
	GetTariffZone(ctx context.Context, distanceKm float64) (*model.TariffZone, error)
}

type pricingRepo struct {
	db *sqlx.DB
}

func NewPricingRepository(db *sqlx.DB) PricingRepository {
	return &pricingRepo{db: db}
}

func (r *pricingRepo) GetServiceByCode(ctx context.Context, code string) (*model.ServiceInfo, error) {
	var svc model.ServiceInfo
	err := r.db.GetContext(ctx, &svc, "SELECT * FROM service_types WHERE code = $1 AND is_active = true", code)
	return &svc, err
}

func (r *pricingRepo) GetAllServices(ctx context.Context) ([]model.ServiceInfo, error) {
	var services []model.ServiceInfo
	err := r.db.SelectContext(ctx, &services, "SELECT * FROM service_types WHERE is_active = true ORDER BY base_fee ASC")
	return services, err
}

func (r *pricingRepo) GetTariffZone(ctx context.Context, distanceKm float64) (*model.TariffZone, error) {
	var zone model.TariffZone
	err := r.db.GetContext(ctx, &zone,
		"SELECT * FROM tariff_zones WHERE min_dist_km <= $1 AND max_dist_km >= $1 LIMIT 1",
		distanceKm)
	if err != nil {
		// Default zone with multiplier 1.0
		return &model.TariffZone{Multiplier: 1.0}, nil
	}
	return &zone, nil
}
