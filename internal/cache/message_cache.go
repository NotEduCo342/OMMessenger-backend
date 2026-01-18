package cache

import (
	"fmt"
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/vmihailenco/msgpack/v5"
)

// TTL constants for different cache types
const (
	ConversationTTL = 5 * time.Minute
	UserListTTL     = 2 * time.Minute
	UnreadCountTTL  = 1 * time.Minute
)

// MessageCache handles message-related caching
type MessageCache struct {
	redis *RedisCache
}

// NewMessageCache creates a new message cache
func NewMessageCache(redis *RedisCache) *MessageCache {
	return &MessageCache{redis: redis}
}

// conversationKey generates a cache key for a conversation
func conversationKey(userID1, userID2 uint) string {
	// Always use smaller ID first for consistency
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	return fmt.Sprintf("conv:%d:%d", userID1, userID2)
}

func groupConversationKey(groupID uint) string {
	return fmt.Sprintf("group:%d", groupID)
}

// GetConversation retrieves cached conversation messages
func (mc *MessageCache) GetConversation(userID1, userID2 uint) ([]models.Message, bool) {
	if mc == nil || mc.redis == nil {
		return nil, false
	}
	key := conversationKey(userID1, userID2)
	data, err := mc.redis.Get(key)
	if err != nil || data == nil {
		return nil, false
	}

	var messages []models.Message
	if err := msgpack.Unmarshal(data, &messages); err != nil {
		return nil, false
	}

	return messages, true
}

// GetGroupConversation retrieves cached group messages
func (mc *MessageCache) GetGroupConversation(groupID uint) ([]models.Message, bool) {
	if mc == nil || mc.redis == nil {
		return nil, false
	}
	key := groupConversationKey(groupID)
	data, err := mc.redis.Get(key)
	if err != nil || data == nil {
		return nil, false
	}

	var messages []models.Message
	if err := msgpack.Unmarshal(data, &messages); err != nil {
		return nil, false
	}

	return messages, true
}

// SetConversation caches conversation messages
func (mc *MessageCache) SetConversation(userID1, userID2 uint, messages []models.Message) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := conversationKey(userID1, userID2)
	data, err := msgpack.Marshal(messages)
	if err != nil {
		return err
	}

	return mc.redis.Set(key, data, ConversationTTL)
}

// SetGroupConversation caches group messages
func (mc *MessageCache) SetGroupConversation(groupID uint, messages []models.Message) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := groupConversationKey(groupID)
	data, err := msgpack.Marshal(messages)
	if err != nil {
		return err
	}

	return mc.redis.Set(key, data, ConversationTTL)
}

// InvalidateConversation removes conversation from cache
func (mc *MessageCache) InvalidateConversation(userID1, userID2 uint) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := conversationKey(userID1, userID2)
	return mc.redis.Delete(key)
}

// InvalidateGroupConversation removes group conversation from cache
func (mc *MessageCache) InvalidateGroupConversation(groupID uint) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := groupConversationKey(groupID)
	return mc.redis.Delete(key)
}

// GetConversationList retrieves cached conversation list for a user
func (mc *MessageCache) GetConversationList(userID uint) ([]interface{}, bool) {
	if mc == nil || mc.redis == nil {
		return nil, false
	}
	key := fmt.Sprintf("convlist:%d", userID)
	data, err := mc.redis.Get(key)
	if err != nil || data == nil {
		return nil, false
	}

	var conversations []interface{}
	if err := msgpack.Unmarshal(data, &conversations); err != nil {
		return nil, false
	}

	return conversations, true
}

// GetConversationListPayload retrieves cached conversation list payload
func (mc *MessageCache) GetConversationListPayload(userID uint) (map[string]interface{}, bool) {
	if mc == nil || mc.redis == nil {
		return nil, false
	}
	key := fmt.Sprintf("convlist:%d", userID)
	data, err := mc.redis.Get(key)
	if err != nil || data == nil {
		return nil, false
	}

	var payload map[string]interface{}
	if err := msgpack.Unmarshal(data, &payload); err != nil {
		return nil, false
	}

	return payload, true
}

// SetConversationList caches conversation list for a user
func (mc *MessageCache) SetConversationList(userID uint, conversations []interface{}) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := fmt.Sprintf("convlist:%d", userID)
	data, err := msgpack.Marshal(conversations)
	if err != nil {
		return err
	}

	return mc.redis.Set(key, data, UserListTTL)
}

// SetConversationListPayload caches conversation list payload
func (mc *MessageCache) SetConversationListPayload(userID uint, payload map[string]interface{}) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := fmt.Sprintf("convlist:%d", userID)
	data, err := msgpack.Marshal(payload)
	if err != nil {
		return err
	}

	return mc.redis.Set(key, data, UserListTTL)
}

// InvalidateConversationList removes conversation list from cache
func (mc *MessageCache) InvalidateConversationList(userID uint) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := fmt.Sprintf("convlist:%d", userID)
	return mc.redis.Delete(key)
}

// GetUnreadCount retrieves cached unread count
func (mc *MessageCache) GetUnreadCount(userID uint, otherUserID uint) (int, bool) {
	if mc == nil || mc.redis == nil {
		return 0, false
	}
	key := fmt.Sprintf("unread:%d:%d", userID, otherUserID)
	data, err := mc.redis.Get(key)
	if err != nil || data == nil {
		return 0, false
	}

	var count int
	if err := msgpack.Unmarshal(data, &count); err != nil {
		return 0, false
	}

	return count, true
}

// SetUnreadCount caches unread count
func (mc *MessageCache) SetUnreadCount(userID uint, otherUserID uint, count int) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := fmt.Sprintf("unread:%d:%d", userID, otherUserID)
	data, err := msgpack.Marshal(count)
	if err != nil {
		return err
	}

	return mc.redis.Set(key, data, UnreadCountTTL)
}

// InvalidateUnreadCount removes unread count from cache
func (mc *MessageCache) InvalidateUnreadCount(userID uint, otherUserID uint) error {
	if mc == nil || mc.redis == nil {
		return nil
	}
	key := fmt.Sprintf("unread:%d:%d", userID, otherUserID)
	return mc.redis.Delete(key)
}
