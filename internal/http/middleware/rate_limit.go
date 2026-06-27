package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/as9840935/url-shortener/internal/metrics"
	"github.com/as9840935/url-shortener/internal/ratelimit"
	"github.com/as9840935/url-shortener/internal/response"
)

type RateLimitConfig struct {
	Type     string
	Limit    int
	Window   time.Duration
	FailOpen bool
}

func RateLimit(limiter ratelimit.Limiter, cfg RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))

			key := fmt.Sprintf("ratelimit:%s:%s", cfg.Type, clientIP(r))
			result, err := limiter.Allow(r.Context(), key, cfg.Limit, cfg.Window)
			if err != nil {
				metrics.RateLimitErrorsTotal.WithLabelValues(cfg.Type).Inc()
				if cfg.FailOpen {
					next.ServeHTTP(w, r)
					return
				}

				response.Error(w, http.StatusServiceUnavailable, "service unavailable")
				return
			}

			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			if !result.Allowed {
				retryAfter := retryAfterSeconds(result.RetryAfter)
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
				metrics.RateLimitedRequestsTotal.WithLabelValues(cfg.Type).Inc()
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		firstIP, _, _ := strings.Cut(forwardedFor, ",")
		if ip := normalizedIP(firstIP); ip != "" {
			return ip
		}
	}

	if ip := normalizedIP(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		if ip := normalizedIP(host); ip != "" {
			return ip
		}
	}
	if ip := normalizedIP(r.RemoteAddr); ip != "" {
		return ip
	}

	return "unknown"
}

func normalizedIP(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	ip := net.ParseIP(value)
	if ip == nil {
		return ""
	}

	return ip.String()
}

func retryAfterSeconds(retryAfter time.Duration) int64 {
	seconds := int64((retryAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		return 1
	}

	return seconds
}
