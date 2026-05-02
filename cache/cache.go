package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

// Cache interface for storing and retrieving SQL queries
type Cache interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key string, value string, ttl time.Duration)
}

// --- Redis Implementation ---

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(uri string) (*RedisCache, error) {
	opts, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) {
	r.client.Set(ctx, key, value, ttl)
}

// --- Memory Implementation ---

type MemoryCache struct {
	store *cache.Cache
}

func NewMemoryCache(defaultExpiration, cleanupInterval time.Duration) *MemoryCache {
	return &MemoryCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (m *MemoryCache) Get(_ context.Context, key string) (string, bool) {
	val, found := m.store.Get(key)
	if !found {
		return "", false
	}
	return val.(string), true
}

func (m *MemoryCache) Set(_ context.Context, key string, value string, ttl time.Duration) {
	m.store.Set(key, value, ttl)
}
