package shorturl

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheContract interface {
	GetOriginalURL(ctx context.Context, code string) (string, error)
	SetOriginalURL(ctx context.Context, code string, originalURL string, ttl time.Duration) error
	IsNotFoundCached(ctx context.Context, code string) (bool, error)
	SetNotFound(ctx context.Context, code string, ttl time.Duration) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) GetOriginalURL(ctx context.Context, code string) (string, error) {
	originalURL, err := c.client.Get(ctx, originalURLCacheKey(code)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}

	return originalURL, err
}

func (c *RedisCache) SetOriginalURL(ctx context.Context, code string, originalURL string, ttl time.Duration) error {
	return c.client.Set(ctx, originalURLCacheKey(code), originalURL, ttl).Err()
}

func (c *RedisCache) IsNotFoundCached(ctx context.Context, code string) (bool, error) {
	count, err := c.client.Exists(ctx, notFoundCacheKey(code)).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (c *RedisCache) SetNotFound(ctx context.Context, code string, ttl time.Duration) error {
	return c.client.Set(ctx, notFoundCacheKey(code), "1", ttl).Err()
}

func originalURLCacheKey(code string) string {
	return "shorturl:" + code
}

func notFoundCacheKey(code string) string {
	return "shorturl:notfound:" + code
}
