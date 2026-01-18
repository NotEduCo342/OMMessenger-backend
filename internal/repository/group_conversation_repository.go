package repository

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// GroupConversationRow is a denormalized row representing a group conversation
// with last message + unread count + group info.
type GroupConversationRow struct {
	GroupID     uint   `gorm:"column:group_id"`
	GroupName   string `gorm:"column:group_name"`
	GroupIcon   string `gorm:"column:group_icon"`
	MemberCount int64  `gorm:"column:member_count"`

	UnreadCount int64 `gorm:"column:unread_count"`

	MessageID          uint      `gorm:"column:message_id"`
	MessageClientID    string    `gorm:"column:message_client_id"`
	MessageSenderID    uint      `gorm:"column:message_sender_id"`
	MessageContent     string    `gorm:"column:message_content"`
	MessageType        string    `gorm:"column:message_type"`
	MessageStatus      string    `gorm:"column:message_status"`
	MessageIsDelivered bool      `gorm:"column:message_is_delivered"`
	MessageIsRead      bool      `gorm:"column:message_is_read"`
	MessageCreatedAt   time.Time `gorm:"column:message_created_at"`

	LastActivity time.Time `gorm:"column:last_activity"`

	SenderID       uint       `gorm:"column:sender_id"`
	SenderUsername string     `gorm:"column:sender_username"`
	SenderEmail    string     `gorm:"column:sender_email"`
	SenderFullName string     `gorm:"column:sender_full_name"`
	SenderAvatar   string     `gorm:"column:sender_avatar"`
	SenderIsOnline bool       `gorm:"column:sender_is_online"`
	SenderLastSeen *time.Time `gorm:"column:sender_last_seen"`
}

func (r *MessageRepository) ListGroupConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]GroupConversationRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	limitPlusOne := limit + 1

	var whereCursor string
	args := []interface{}{userID, userID}
	if cursorCreatedAt != nil && cursorMessageID > 0 {
		whereCursor = "AND (t.message_created_at < ? OR (t.message_created_at = ? AND t.message_id < ?))"
		args = append(args, *cursorCreatedAt, *cursorCreatedAt, cursorMessageID)
	}
	args = append(args, limitPlusOne)

	query := strings.TrimSpace(`
WITH ranked AS (
	SELECT
		m.group_id AS group_id,
		m.id AS message_id,
		m.client_id AS message_client_id,
		m.sender_id AS message_sender_id,
		m.content AS message_content,
		m.message_type AS message_type,
		m.status AS message_status,
		m.is_delivered AS message_is_delivered,
		m.is_read AS message_is_read,
		m.created_at AS message_created_at,
		m.created_at AS last_activity,
		ROW_NUMBER() OVER (
			PARTITION BY m.group_id
			ORDER BY m.created_at DESC, m.id DESC
		) AS rn,
		SUM(CASE WHEN m.id > COALESCE(grs.last_read_message_id, 0) THEN 1 ELSE 0 END) OVER (
			PARTITION BY m.group_id
		) AS unread_count
	FROM messages m
	JOIN group_members gm ON gm.group_id = m.group_id AND gm.user_id = ?
	LEFT JOIN group_read_states grs ON grs.group_id = m.group_id AND grs.user_id = ?
	WHERE m.group_id IS NOT NULL
)
SELECT
	t.group_id,
	g.name AS group_name,
	g.icon AS group_icon,
	(
		SELECT COUNT(*)
		FROM group_members gm2
		WHERE gm2.group_id = t.group_id
	) AS member_count,
	t.unread_count,
	t.message_id,
	t.message_client_id,
	t.message_sender_id,
	t.message_content,
	t.message_type,
	t.message_status,
	t.message_is_delivered,
	t.message_is_read,
	t.message_created_at,
	t.last_activity,
	sender.id AS sender_id,
	sender.username AS sender_username,
	sender.email AS sender_email,
	sender.full_name AS sender_full_name,
	sender.avatar AS sender_avatar,
	sender.is_online AS sender_is_online,
	sender.last_seen AS sender_last_seen
FROM ranked t
JOIN groups g ON g.id = t.group_id
JOIN users sender ON sender.id = t.message_sender_id
WHERE t.rn = 1
` + "\n" + whereCursor + `
ORDER BY t.last_activity DESC, t.message_id DESC
LIMIT ?
`)

	var rows []GroupConversationRow
	if err := r.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

var _ = gorm.ErrRecordNotFound
