package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService provides Redis-based caching operations
type CacheService struct {
	client *redis.Client
}

// NewCacheService creates a new cache service with the given Redis client
func NewCacheService(client *redis.Client) *CacheService {
	return &CacheService{client: client}
}

// Cache key prefixes
const (
	PrefixGroupStats = "stats:group:"
	PrefixUserStats  = "stats:user:"
	PrefixGroup      = "group:"
	PrefixBills      = "bills:group:"
	PrefixBalances   = "balances:group:"
	PrefixCategories = "categories"
)

// Default TTLs
const (
	TTLGroupStats = 5 * time.Minute
	TTLUserStats  = 5 * time.Minute
	TTLGroup      = 10 * time.Minute
	TTLBills      = 3 * time.Minute
	TTLBalances   = 2 * time.Minute
	TTLCategories = 24 * time.Hour
)

// Get retrieves a cached value by key and unmarshals it into the target
func (c *CacheService) Get(ctx context.Context, key string, target interface{}) error {
	if c.client == nil {
		return fmt.Errorf("redis client not available")
	}

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), target)
}

// Set caches a value with the given key and TTL
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("redis client not available")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a cached value by key
func (c *CacheService) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, key).Err()
}

// DeletePattern removes all cached values matching a pattern
func (c *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	if c.client == nil {
		return nil
	}

	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// InvalidateGroupCache invalidates all caches related to a group
func (c *CacheService) InvalidateGroupCache(ctx context.Context, groupID string) error {
	keys := []string{
		PrefixGroup + groupID,
		PrefixGroupStats + groupID,
		PrefixBills + groupID,
		PrefixBalances + groupID,
	}

	for _, key := range keys {
		_ = c.Delete(ctx, key)
	}
	return nil
}

// InvalidateUserCache invalidates all caches related to a user
func (c *CacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	return c.Delete(ctx, PrefixUserStats+userID)
}

// GetOrSet attempts to get from cache; on miss, calls fn and caches the result
func (c *CacheService) GetOrSet(ctx context.Context, key string, target interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	// Try cache first
	err := c.Get(ctx, key, target)
	if err == nil {
		return nil // Cache hit
	}

	// Cache miss - call the function
	result, err := fn()
	if err != nil {
		return err
	}

	// Cache the result (best effort)
	_ = c.Set(ctx, key, result, ttl)

	// Marshal and unmarshal to populate target
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// IsAvailable checks if the Redis client is connected and available
func (c *CacheService) IsAvailable() bool {
	if c.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.client.Ping(ctx).Err() == nil
}
