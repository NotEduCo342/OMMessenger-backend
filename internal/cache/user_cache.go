package cache

import (
	"fmt"
	"strconv"
	"time"
)

const (
	OnlineUsersTTL = 90 * time.Second // Match pong timeout
)

// UserCache handles user-related caching
type UserCache struct {
	redis *RedisCache
}

// NewUserCache creates a new user cache
func NewUserCache(redis *RedisCache) *UserCache {
	return &UserCache{redis: redis}
}

// SetUserOnline adds a user to the online users set
func (uc *UserCache) SetUserOnline(userID uint) error {
	if uc == nil || uc.redis == nil {
		return nil
	}
	key := "online:users"
	if err := uc.redis.SetAdd(key, userID); err != nil {
		return err
	}

	// Set individual user key with TTL for auto-expiration
	userKey := fmt.Sprintf("online:%d", userID)
	return uc.redis.Set(userKey, []byte("1"), OnlineUsersTTL)
}

// SetUserOffline removes a user from the online users set
func (uc *UserCache) SetUserOffline(userID uint) error {
	if uc == nil || uc.redis == nil {
		return nil
	}
	key := "online:users"
	if err := uc.redis.SetRemove(key, userID); err != nil {
		return err
	}

	// Delete individual user key
	userKey := fmt.Sprintf("online:%d", userID)
	return uc.redis.Delete(userKey)
}

// IsUserOnline checks if a user is online
func (uc *UserCache) IsUserOnline(userID uint) bool {
	if uc == nil || uc.redis == nil {
		return false
	}
	userKey := fmt.Sprintf("online:%d", userID)
	return uc.redis.Exists(userKey)
}

// GetOnlineUsers returns all online user IDs
func (uc *UserCache) GetOnlineUsers() ([]uint, error) {
	if uc == nil || uc.redis == nil {
		return nil, nil
	}
	key := "online:users"
	members, err := uc.redis.SetMembers(key)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0, len(members))
	for _, member := range members {
		if id, err := strconv.ParseUint(member, 10, 32); err == nil {
			userIDs = append(userIDs, uint(id))
		}
	}

	return userIDs, nil
}

// GetOnlineCount returns the number of online users
func (uc *UserCache) GetOnlineCount() (int64, error) {
	if uc == nil || uc.redis == nil {
		return 0, nil
	}
	key := "online:users"
	return uc.redis.SetCard(key)
}

// RefreshUserOnline extends the TTL for an online user
func (uc *UserCache) RefreshUserOnline(userID uint) error {
	if uc == nil || uc.redis == nil {
		return nil
	}
	userKey := fmt.Sprintf("online:%d", userID)
	return uc.redis.Set(userKey, []byte("1"), OnlineUsersTTL)
}
