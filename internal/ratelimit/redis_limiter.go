package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client redis.Scripter
	script *redis.Script
}

func NewRedisLimiter(client redis.Scripter) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		script: redis.NewScript(tokenBucketLua),
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	if key == "" {
		return Result{}, fmt.Errorf("rate limit key is required")
	}
	if limit <= 0 {
		return Result{}, fmt.Errorf("rate limit must be greater than zero")
	}
	if window.Milliseconds() <= 0 {
		return Result{}, fmt.Errorf("rate limit window must be at least one millisecond")
	}

	values, err := l.script.Run(ctx, l.client, []string{key}, limit, window.Milliseconds()).Slice()
	if err != nil {
		return Result{}, fmt.Errorf("run rate limit script: %w", err)
	}
	if len(values) != 3 {
		return Result{}, fmt.Errorf("rate limit script returned %d values", len(values))
	}

	allowed, err := redisValueAsInt64(values[0])
	if err != nil {
		return Result{}, fmt.Errorf("decode allowed: %w", err)
	}
	remaining, err := redisValueAsInt64(values[1])
	if err != nil {
		return Result{}, fmt.Errorf("decode remaining: %w", err)
	}
	retryAfterMilliseconds, err := redisValueAsInt64(values[2])
	if err != nil {
		return Result{}, fmt.Errorf("decode retry after: %w", err)
	}

	return Result{
		Allowed:    allowed == 1,
		Limit:      limit,
		Remaining:  int(remaining),
		RetryAfter: time.Duration(retryAfterMilliseconds) * time.Millisecond,
	}, nil
}

func redisValueAsInt64(value any) (int64, error) {
	switch typedValue := value.(type) {
	case int64:
		return typedValue, nil
	case string:
		return strconv.ParseInt(typedValue, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected type %T", value)
	}
}
