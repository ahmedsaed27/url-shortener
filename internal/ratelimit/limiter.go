package ratelimit

import (
	"context"
	"time"
)

// Limiter decides whether a key may perform an action within a window.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error)
}
