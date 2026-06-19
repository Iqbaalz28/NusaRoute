// Package hubs resolves the nearest sortation hub to a coordinate by reading the
// hub-service hub list (cached). Used to stamp an order's origin/destination hub.
package hubs

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Code string  `json:"code"`
	City string  `json:"city"`
	Lat  float64  `json:"lat"`
	Lng  float64  `json:"lng"`
}

// Resolver fetches and caches the hub list, refreshing at most once per TTL.
type Resolver struct {
	baseURL string
	ttl     time.Duration
	client  *http.Client

	mu       sync.Mutex
	cache    []Hub
	fetched  time.Time
}

func NewResolver(baseURL string) *Resolver {
	return &Resolver{
		baseURL: baseURL,
		ttl:     5 * time.Minute,
		client:  &http.Client{Timeout: 3 * time.Second},
	}
}

func (r *Resolver) hubs(ctx context.Context) []Hub {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cache != nil && time.Since(r.fetched) < r.ttl {
		return r.cache
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.baseURL+"/api/v1/hub/list", nil)
	if err != nil {
		return r.cache
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return r.cache // serve stale on failure
	}
	defer resp.Body.Close()

	var env struct {
		Data []Hub `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return r.cache
	}
	r.cache = env.Data
	r.fetched = time.Now()
	return r.cache
}

// Nearest returns the hub closest to (lat,lng). Ok is false when no hub can be
// resolved (empty list or zero coordinates).
func (r *Resolver) Nearest(ctx context.Context, lat, lng float64) (Hub, bool) {
	if lat == 0 && lng == 0 {
		return Hub{}, false
	}
	hubs := r.hubs(ctx)
	if len(hubs) == 0 {
		return Hub{}, false
	}
	best := hubs[0]
	bestDist := haversineKm(lat, lng, best.Lat, best.Lng)
	for _, h := range hubs[1:] {
		if d := haversineKm(lat, lng, h.Lat, h.Lng); d < bestDist {
			best, bestDist = h, d
		}
	}
	return best, true
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
