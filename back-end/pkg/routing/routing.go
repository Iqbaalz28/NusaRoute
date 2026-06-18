// Package routing decides how a package should move from sender to receiver:
// delivered DIRECTLY by the pickup courier (same city + instant service) or
// routed VIA a sortation HUB (everything else).
package routing

import (
	"math"
	"strings"
)

const (
	// RouteDirect: courier delivers straight to the receiver, no hub leg.
	RouteDirect = "DIRECT"
	// RouteViaHub: package must pass through a sortation hub first.
	RouteViaHub = "VIA_HUB"

	// SameCityThresholdKm is the max sender→receiver distance still considered
	// "intra-city" (so it can skip the hub for instant deliveries).
	SameCityThresholdKm = 30.0
)

// IsInstantService reports whether a service type is same-day / instant class.
func IsInstantService(serviceType string) bool {
	switch strings.ToUpper(strings.TrimSpace(serviceType)) {
	case "SAME", "SAMEDAY", "SAME_DAY", "INSTANT":
		return true
	default:
		return false
	}
}

// HaversineKm returns the great-circle distance in kilometers between two points.
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// IsSameCity reports whether a distance is within the intra-city threshold.
func IsSameCity(distanceKm float64) bool {
	return distanceKm <= SameCityThresholdKm
}

// DecideRoute returns RouteDirect only when an instant-class service is being
// shipped within the same city; otherwise the package goes VIA_HUB.
func DecideRoute(serviceType string, distanceKm float64) string {
	if IsInstantService(serviceType) && IsSameCity(distanceKm) {
		return RouteDirect
	}
	return RouteViaHub
}

// DecideRouteByCoords is DecideRoute computed from sender/receiver coordinates.
// Missing coordinates (0,0) yield a large distance, which safely falls back to VIA_HUB.
func DecideRouteByCoords(serviceType string, senderLat, senderLng, receiverLat, receiverLng float64) string {
	if senderLat == 0 && senderLng == 0 || receiverLat == 0 && receiverLng == 0 {
		// Unknown location → cannot guarantee same-city → use the hub.
		return RouteViaHub
	}
	return DecideRoute(serviceType, HaversineKm(senderLat, senderLng, receiverLat, receiverLng))
}
