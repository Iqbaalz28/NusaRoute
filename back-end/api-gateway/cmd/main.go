package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"context"
	"fmt"
	"github.com/nusaroute/pkg/logger"
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
	logger.InitLogger("api-gateway")
		srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), fmt.Sprint("Shutting down service gracefully..."))
		cancel() // notify workers to stop

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	logger.Info(context.Background(), fmt.Sprintf("✅ Service running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), fmt.Sprint("All workers stopped. Goodbye!"))

}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
