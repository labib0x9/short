package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labib0x9/short/internal/domain/cache"
	"github.com/labib0x9/short/internal/utils"
)

type RateLimiter struct {
	Client   cache.RateLimiter
	Rate     int
	Capacity int
}

type rateLimitResult struct {
	allowed     bool
	wait_ms     int64
	token       int
	last_refill int64
}

func NewRateLimiter(
	client cache.RateLimiter,
	rate int,
	capacity int,
) *RateLimiter {
	return &RateLimiter{
		Client:   client,
		Rate:     rate,
		Capacity: capacity,
	}
}

func (rl *RateLimiter) Limit() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, err := getIP(r)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			key := "rate_limit:ip:" + ip + ":path:" + r.URL.Path
			res, err := rl.setLimit(r.Context(), key)
			if err != nil {
				utils.SendError(w, "internal server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.Rate))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.token))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(res.last_refill)))

			if !res.allowed {
				retryAfterSecs := res.wait_ms / 1000
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSecs, 10))
				utils.SendError(w, "too many request", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) setLimit(ctx context.Context, key string) (rateLimitResult, error) {
	now := time.Now().UnixMilli()

	res, err := rl.Client.RunScript(ctx, key, rl.Capacity, rl.Rate, now)
	if err != nil {
		return rateLimitResult{}, err
	}

	data := res.([]interface{})
	allowed := data[0].(int64)
	wait_ms := data[1].(int64)
	token := data[2].(int64)

	return rateLimitResult{
		allowed:     allowed == 1,
		wait_ms:     wait_ms,
		last_refill: now,
		token:       int(token),
	}, nil
}

func getIP(r *http.Request) (string, error) {
	addr := r.RemoteAddr
	host, _, err := net.SplitHostPort(addr)
	return host, err
}
