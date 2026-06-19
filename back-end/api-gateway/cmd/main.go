package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nusaroute/pkg/database"
	"github.com/nusaroute/pkg/logger"
	mw "github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/pkg/response"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// ServiceRoute maps URL prefix to backend service URL.
type ServiceRoute struct {
	Prefix    string
	TargetURL string
	Public    bool     // if true, skip JWT validation
	Roles     []string // if set, require one of these roles (implies JWT)
}

// RedisRateLimiter implements a fixed-window rate limiter using Redis.
type RedisRateLimiter struct {
	client *redis.Client
	rate   int // allowed requests per window
	window time.Duration
}

func NewRedisRateLimiter(client *redis.Client, rate int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		rate:   rate,
		window: window,
	}
}

func (rl *RedisRateLimiter) Allow(ip string) bool {
	if rl.client == nil {
		return true // Fallback if redis is down
	}
	windowKey := time.Now().Unix() / int64(rl.window.Seconds())
	key := fmt.Sprintf("ratelimit:%s:%d", ip, windowKey)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Rate limiter Redis error: %v", err))
		return true // fail-open
	}

	if count == 1 {
		rl.client.Expire(ctx, key, rl.window)
	}

	return count <= int64(rl.rate)
}

func main() {
	logger.InitLogger("api-gateway")
	logger.Info(context.Background(), "Starting NusaRoute API Gateway...")
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "nusaroute-jwt-secret-key-2026")

	appEnv := getEnv("APP_ENV", "development")
	if appEnv == "production" && jwtSecret == "nusaroute-jwt-secret-key-2026" {
		logger.Log.Fatal("FATAL: JWT_SECRET must be set securely in production environment!")
	}

	// Define service routes
	routes := []ServiceRoute{
		{Prefix: "/api/v1/users", TargetURL: getEnv("USER_SERVICE_URL", "http://localhost:8001"), Public: false},
		{Prefix: "/api/v1/auth", TargetURL: getEnv("USER_SERVICE_URL", "http://localhost:8001"), Public: true},
		{Prefix: "/api/v1/payments", TargetURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8002"), Public: false},
		{Prefix: "/api/v1/pricing", TargetURL: getEnv("PRICING_SERVICE_URL", "http://localhost:8003"), Public: true},
		{Prefix: "/api/v1/orders", TargetURL: getEnv("ORDER_SERVICE_URL", "http://localhost:8004"), Public: false},
		{Prefix: "/api/v1/couriers", TargetURL: getEnv("COURIER_SERVICE_URL", "http://localhost:8005"), Public: false},
		{Prefix: "/api/v1/dispatch", TargetURL: getEnv("DISPATCH_SERVICE_URL", "http://localhost:8006"), Public: false, Roles: []string{"COURIER", "ADMIN"}},
		{Prefix: "/api/v1/hub", TargetURL: getEnv("HUB_SERVICE_URL", "http://localhost:8007"), Public: true},
		{Prefix: "/api/v1/tracking", TargetURL: getEnv("TRACKING_SERVICE_URL", "http://localhost:8008"), Public: true},
		{Prefix: "/api/v1/notifications", TargetURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8009"), Public: false},
		{Prefix: "/api/v1/resolutions", TargetURL: getEnv("RESOLUTION_SERVICE_URL", "http://localhost:8010"), Public: false},
	}

	redisClient, err := database.ConnectRedis(database.RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
	})
	if err != nil {
		logger.Info(context.Background(), fmt.Sprintf("Redis unavailable, rate limiting will be bypassed: %v", err))
	}
	rateLimiter := NewRedisRateLimiter(redisClient, 100, time.Minute)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"api-gateway"}`))
	})

	// Custom Dashboard Routes (Aggregation)
	mux.Handle("/api/v1/dashboard/stats", mw.JWTAuth(jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var orderStats, courierStats, hubStats map[string]interface{}

		client := &http.Client{Timeout: 5 * time.Second}

		wg.Add(3)
		// Fetch Order Stats
		go func() {
			defer wg.Done()
			resp, err := client.Get(getEnv("ORDER_SERVICE_URL", "http://localhost:8004") + "/api/v1/orders/stats")
			if err == nil {
				defer resp.Body.Close()
			}
			if err == nil && resp.StatusCode == http.StatusOK {
				var res struct {
					Data map[string]interface{} `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&res)
				orderStats = res.Data
			}
		}()

		// Fetch Courier Stats
		go func() {
			defer wg.Done()
			resp, err := client.Get(getEnv("COURIER_SERVICE_URL", "http://localhost:8005") + "/api/v1/couriers/stats")
			if err == nil {
				defer resp.Body.Close()
			}
			if err == nil && resp.StatusCode == http.StatusOK {
				var res struct {
					Data map[string]interface{} `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&res)
				courierStats = res.Data
			}
		}()

		// Fetch Hub Stats
		go func() {
			defer wg.Done()
			resp, err := client.Get(getEnv("HUB_SERVICE_URL", "http://localhost:8007") + "/api/v1/hub/stats")
			if err == nil {
				defer resp.Body.Close()
			}
			if err == nil && resp.StatusCode == http.StatusOK {
				var res struct {
					Data map[string]interface{} `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&res)
				hubStats = res.Data
			}
		}()

		wg.Wait()

		stats := map[string]interface{}{
			"total_orders_today": 0,
			"active_couriers":    0,
			"active_hubs":        0,
			"sla_percentage":     0.0,
			"total_cities":       0,
		}

		if orderStats != nil {
			stats["total_orders_today"] = orderStats["total_orders_today"]
			stats["sla_percentage"] = orderStats["sla_percentage"]
		}
		if courierStats != nil {
			stats["active_couriers"] = courierStats["active_couriers"]
		}
		if hubStats != nil {
			stats["active_hubs"] = hubStats["active_hubs"]
			stats["total_cities"] = hubStats["total_cities"]
		}

		response.Success(w, "dashboard stats retrieved", stats)
	})))

	mux.Handle("/api/v1/dashboard/volume", mw.JWTAuth(jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(getEnv("ORDER_SERVICE_URL", "http://localhost:8004") + "/api/v1/orders/volume")
		if err != nil {
			response.InternalError(w, err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			response.InternalError(w, "failed to fetch volume stats")
			return
		}

		var res struct {
			Data interface{} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			response.InternalError(w, err.Error())
			return
		}

		response.Success(w, "dashboard volume retrieved", res.Data)
	})))

	// makeProxyHandler builds a rate-limited reverse-proxy handler for a target
	// service URL. Shared by the generic route loop and by routes that need a
	// different auth policy than their parent prefix (e.g. hub scan).
	makeProxyHandler := func(targetURL string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rate limiting
			ip := strings.Split(r.RemoteAddr, ":")[0]
			if !rateLimiter.Allow(ip) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			target, err := url.Parse(targetURL)
			if err != nil {
				http.Error(w, `{"error":"bad gateway"}`, http.StatusBadGateway)
				return
			}

			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.Transport = &http.Transport{
				ResponseHeaderTimeout: 10 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   2 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			}
			proxy.ModifyResponse = func(resp *http.Response) error {
				resp.Header.Del("Access-Control-Allow-Origin")
				resp.Header.Del("Access-Control-Allow-Methods")
				resp.Header.Del("Access-Control-Allow-Headers")
				return nil
			}
			proxy.ServeHTTP(w, r)
		})
	}

	// Hub scan endpoints require a hub operator (or admin) even though the rest of
	// /api/v1/hub (list, stats, manifest) stays public for the dashboard.
	// Registered as a more specific prefix so ServeMux longest-prefix match routes
	// scans through JWT + role enforcement.
	mux.Handle("/api/v1/hub/scan/", mw.JWTAuth(jwtSecret)(mw.RequireRole("HUB_OPERATOR", "ADMIN")(
		makeProxyHandler(getEnv("HUB_SERVICE_URL", "http://localhost:8007")))))

	// Admin-only order endpoints. These are more specific than the generic
	// /api/v1/orders prefix (which customers use to create/list/cancel their own
	// orders), so ServeMux routes them here for ADMIN-only enforcement.
	adminOrders := mw.JWTAuth(jwtSecret)(mw.RequireRole("ADMIN")(
		makeProxyHandler(getEnv("ORDER_SERVICE_URL", "http://localhost:8004"))))
	mux.Handle("/api/v1/orders/all", adminOrders)
	mux.Handle("/api/v1/orders/status", adminOrders)

	for _, route := range routes {
		route := route // capture loop var
		proxyHandler := makeProxyHandler(route.TargetURL)

		// Role check runs inside JWT auth so the role is already in context.
		if len(route.Roles) > 0 {
			proxyHandler = mw.RequireRole(route.Roles...)(proxyHandler)
		}

		// Apply JWT Auth only if the route is not public (role check implies JWT).
		if !route.Public || len(route.Roles) > 0 {
			proxyHandler = mw.JWTAuth(jwtSecret)(proxyHandler)
		}

		mux.Handle(route.Prefix+"/", proxyHandler)
		mux.Handle(route.Prefix, proxyHandler) // Daftarkan juga rute tanpa slash
	}

	// Apply global middleware chain
	var h http.Handler = mux
	h = mw.CORS(h)
	// JWTAuth is now applied per-route above

	h = mw.Logging(h)
	h = mw.Recovery(h)
	h = mw.Metrics(h)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	_ = ctx // reserved for future background workers

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), "Shutting down API Gateway gracefully...")
		cancel()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	logger.Info(context.Background(), fmt.Sprintf("API Gateway running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), "All workers stopped. Goodbye!")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
