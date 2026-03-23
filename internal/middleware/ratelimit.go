package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/betpro/server/pkg/utils"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client        *redis.Client
	trustProxyHops int
}

func NewRateLimiter(client *redis.Client, trustProxyHops int) *RateLimiter {
	return &RateLimiter{
		client:        client,
		trustProxyHops: trustProxyHops,
	}
}

func (rl *RateLimiter) Limit(action string, maxAttempts int, windowSec int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl.client == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := utils.GetClientIP(r, rl.trustProxyHops)
			key := fmt.Sprintf("ratelimit:%s:%s", action, ip)

			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			count, err := rl.client.Incr(ctx, key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rl.client.Expire(ctx, key, time.Duration(windowSec)*time.Second)
			}

			if count > int64(maxAttempts) {
				respondJSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
