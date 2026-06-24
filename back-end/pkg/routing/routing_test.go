package routing

import (
	"math"
	"testing"
)

func TestIsInstantService(t *testing.T) {
	cases := map[string]bool{
		"SAME": true, "SAMEDAY": true, "same_day": true, "Instant": true, " same ": true,
		"REG": false, "REGULAR": false, "YES": false, "CARGO": false, "": false,
	}
	for in, want := range cases {
		if got := IsInstantService(in); got != want {
			t.Errorf("IsInstantService(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestHaversineKm(t *testing.T) {
	// Jakarta (-6.2088,106.8456) -> Bandung (-6.9175,107.6191) ≈ 117 km
	d := HaversineKm(-6.2088, 106.8456, -6.9175, 107.6191)
	if math.Abs(d-117) > 5 {
		t.Errorf("Jakarta→Bandung = %.1f km, want ≈117", d)
	}
	// Same point = 0
	if d0 := HaversineKm(-6.2, 106.8, -6.2, 106.8); d0 != 0 {
		t.Errorf("same point = %.4f, want 0", d0)
	}
}

func TestDecideRoute(t *testing.T) {
	cases := []struct {
		name        string
		serviceType string
		distanceKm  float64
		want        string
	}{
		{"instant same-city → direct", "SAMEDAY", 8, RouteDirect},
		{"instant edge of city → direct", "INSTANT", SameCityThresholdKm, RouteDirect},
		{"instant just outside city → via hub", "SAMEDAY", SameCityThresholdKm + 0.1, RouteViaHub},
		{"instant intercity → via hub", "SAME", 117, RouteViaHub},
		{"regular same-city → via hub", "REGULAR", 5, RouteViaHub},
		{"regular intercity → via hub", "REG", 117, RouteViaHub},
		{"express same-city → via hub", "EXPRESS", 5, RouteViaHub},
		{"cargo → via hub", "CARGO", 2, RouteViaHub},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DecideRoute(c.serviceType, c.distanceKm); got != c.want {
				t.Errorf("DecideRoute(%q, %.1f) = %s, want %s", c.serviceType, c.distanceKm, got, c.want)
			}
		})
	}
}

func TestDecideRouteByCoords(t *testing.T) {
	// Instant within Jakarta (~5km) → DIRECT
	if got := DecideRouteByCoords("SAMEDAY", -6.2088, 106.8456, -6.1751, 106.8650); got != RouteDirect {
		t.Errorf("intra-Jakarta instant = %s, want DIRECT", got)
	}
	// Instant Jakarta→Bandung (~117km) → VIA_HUB
	if got := DecideRouteByCoords("SAMEDAY", -6.2088, 106.8456, -6.9175, 107.6191); got != RouteViaHub {
		t.Errorf("Jakarta→Bandung instant = %s, want VIA_HUB", got)
	}
	// Missing coordinates → VIA_HUB (safe fallback)
	if got := DecideRouteByCoords("SAMEDAY", 0, 0, -6.9, 107.6); got != RouteViaHub {
		t.Errorf("missing sender coords = %s, want VIA_HUB", got)
	}
	// Regular intra-city → VIA_HUB
	if got := DecideRouteByCoords("REGULAR", -6.2088, 106.8456, -6.1751, 106.8650); got != RouteViaHub {
		t.Errorf("intra-Jakarta regular = %s, want VIA_HUB", got)
	}
}
