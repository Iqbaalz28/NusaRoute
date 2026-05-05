package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	mw "github.com/nusaroute/pkg/middleware"
)

// ServiceRoute maps URL prefix to backend service URL.
type ServiceRoute struct {
	Prefix    string
	TargetURL string
	Public    bool // if true, skip JWT validation
}

// RateLimiter implements a simple token bucket rate limiter per IP.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // tokens per interval
	interval time.Duration
}

type visitor struct {
	tokens   int
	lastSeen time.Time
}

func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
	}
	// Cleanup stale entries every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{tokens: rl.rate - 1, lastSeen: time.Now()}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := time.Since(v.lastSeen)
	refill := int(elapsed / rl.interval) * rl.rate
	v.tokens += refill
	if v.tokens > rl.rate {
		v.tokens = rl.rate
	}
	v.lastSeen = time.Now()

	if v.tokens <= 0 {
		return false
	}
	v.tokens--
	return true
}

func main() {
	log.Println("🚀 Starting NusaRoute API Gateway...")
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "nusaroute-jwt-secret-key-2026")

	// Define service routes
	routes := []ServiceRoute{
		// Public routes (no JWT required)
		{Prefix: "/api/v1/auth/", TargetURL: getEnv("USER_SERVICE_URL", "http://localhost:8001"), Public: true},
		{Prefix: "/api/v1/pricing/", TargetURL: getEnv("PRICING_SERVICE_URL", "http://localhost:8003"), Public: true},
		{Prefix: "/api/v1/tracking/", TargetURL: getEnv("TRACKING_SERVICE_URL", "http://localhost:8008"), Public: true},
		{Prefix: "/api/v1/payments/webhook", TargetURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8002"), Public: true},

		// Protected routes (JWT required)
		{Prefix: "/api/v1/users/", TargetURL: getEnv("USER_SERVICE_URL", "http://localhost:8001"), Public: false},
		{Prefix: "/api/v1/payments/", TargetURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8002"), Public: false},
		{Prefix: "/api/v1/orders/", TargetURL: getEnv("ORDER_SERVICE_URL", "http://localhost:8004"), Public: false},
		{Prefix: "/api/v1/couriers/", TargetURL: getEnv("COURIER_SERVICE_URL", "http://localhost:8005"), Public: false},
		{Prefix: "/api/v1/dispatch/", TargetURL: getEnv("DISPATCH_SERVICE_URL", "http://localhost:8006"), Public: false},
		{Prefix: "/api/v1/hub/", TargetURL: getEnv("HUB_SERVICE_URL", "http://localhost:8007"), Public: false},
		{Prefix: "/api/v1/notifications/", TargetURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8009"), Public: false},
		{Prefix: "/api/v1/resolutions/", TargetURL: getEnv("RESOLUTION_SERVICE_URL", "http://localhost:8010"), Public: false},
	}

	// Rate limiter: 100 requests per second per IP
	rateLimiter := NewRateLimiter(100, time.Second)

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"UP","service":"api-gateway","version":"1.0.0"}`))
	})

	// Register reverse proxy routes
	for _, route := range routes {
		route := route // capture
		target, err := url.Parse(route.TargetURL)
		if err != nil {
			log.Fatalf("Invalid URL for %s: %v", route.Prefix, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		// Custom error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[Gateway] ❌ Proxy error for %s: %v", r.URL.Path, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"success":false,"error":"service unavailable"}`))
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rate limiting
			ip := strings.Split(r.RemoteAddr, ":")[0]
			if !rateLimiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success":false,"error":"rate limit exceeded"}`))
				return
			}

			// JWT validation for protected routes
			if !route.Public {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"success":false,"error":"authorization required"}`))
					return
				}

				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"success":false,"error":"invalid authorization format"}`))
					return
				}

				claims := &mw.JWTClaims{}
				_, err := mw.ParseToken(parts[1], jwtSecret, claims)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"success":false,"error":"invalid or expired token"}`))
					return
				}

				// Forward user info to downstream services via headers
				r.Header.Set("X-User-ID", claims.UserID)
				r.Header.Set("X-User-Role", claims.Role)
				r.Header.Set("X-User-Email", claims.Email)
			}

			proxy.ServeHTTP(w, r)
		})

		mux.Handle(route.Prefix, handler)
	}

	// Apply middleware
	var h http.Handler = mux
	h = mw.CORS(h)
	h = mw.Logging(h)
	h = mw.Recovery(h)

	log.Printf("✅ API Gateway running on port %s", port)
	log.Println("   Routes:")
	for _, r := range routes {
		visibility := "🔒"
		if r.Public { visibility = "🌐" }
		log.Printf("   %s %s → %s", visibility, r.Prefix, r.TargetURL)
	}

	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
