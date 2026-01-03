package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache wraps the Redis client with common operations
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr, password string, db int) *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ctx: context.Background(),
	}
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(key string) ([]byte, error) {
	val, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Key doesn't exist
	}
	return val, err
}

// Set stores a value in Redis with TTL
func (c *RedisCache) Set(key string, value []byte, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

// Delete removes a key from Redis
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *RedisCache) DeletePattern(pattern string) error {
	iter := c.client.Scan(c.ctx, 0, pattern, 0).Iterator()
	for iter.Next(c.ctx) {
		if err := c.client.Del(c.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Exists checks if a key exists
func (c *RedisCache) Exists(key string) bool {
	count, _ := c.client.Exists(c.ctx, key).Result()
	return count > 0
}

// SetAdd adds members to a Redis set
func (c *RedisCache) SetAdd(key string, members ...interface{}) error {
	return c.client.SAdd(c.ctx, key, members...).Err()
}

// SetRemove removes members from a Redis set
func (c *RedisCache) SetRemove(key string, members ...interface{}) error {
	return c.client.SRem(c.ctx, key, members...).Err()
}

// SetMembers returns all members of a Redis set
func (c *RedisCache) SetMembers(key string) ([]string, error) {
	return c.client.SMembers(c.ctx, key).Result()
}

// SetIsMember checks if a value is a member of a set
func (c *RedisCache) SetIsMember(key string, member interface{}) bool {
	isMember, _ := c.client.SIsMember(c.ctx, key, member).Result()
	return isMember
}

// SetCard returns the number of members in a set
func (c *RedisCache) SetCard(key string) (int64, error) {
	return c.client.SCard(c.ctx, key).Result()
}

// Ping checks if Redis is alive
func (c *RedisCache) Ping() error {
	return c.client.Ping(c.ctx).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
