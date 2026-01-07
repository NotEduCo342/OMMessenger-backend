package repository

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// ConversationRow is a denormalized row representing a single direct-message conversation
// (1 row per peer) with last message + unread count + peer profile.
//
// NOTE: This is deliberately not the full models.User / models.Message shape to avoid
// leaking sensitive fields (e.g., peer email) and to keep the query efficient.
type ConversationRow struct {
	PeerID       uint       `gorm:"column:peer_id"`
	PeerUsername string     `gorm:"column:peer_username"`
	PeerEmail    string     `gorm:"column:peer_email"`
	PeerFullName string     `gorm:"column:peer_full_name"`
	PeerAvatar   string     `gorm:"column:peer_avatar"`
	PeerIsOnline bool       `gorm:"column:peer_is_online"`
	PeerLastSeen *time.Time `gorm:"column:peer_last_seen"`

	UnreadCount int64 `gorm:"column:unread_count"`

	MessageID          uint      `gorm:"column:message_id"`
	MessageClientID    string    `gorm:"column:message_client_id"`
	MessageSenderID    uint      `gorm:"column:message_sender_id"`
	MessageRecipientID *uint     `gorm:"column:message_recipient_id"`
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

func (r *MessageRepository) ListDirectConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]ConversationRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// We request one extra row to determine whether there is a next page.
	limitPlusOne := limit + 1

	var whereCursor string
	args := []interface{}{userID, userID, userID, userID, userID, userID}
	if cursorCreatedAt != nil && cursorMessageID > 0 {
		whereCursor = "AND (t.message_created_at < ? OR (t.message_created_at = ? AND t.message_id < ?))"
		args = append(args, *cursorCreatedAt, *cursorCreatedAt, cursorMessageID)
	}
	args = append(args, limitPlusOne)

	// Best-practice notes:
	// - single query, no N+1
	// - window functions pick latest message per peer and compute unread_count per peer
	// - user profiles joined for both peer and last-message sender
	// - excludes group messages
	query := strings.TrimSpace(`
WITH ranked AS (
	SELECT
		CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END AS peer_id,
		m.id AS message_id,
		m.client_id AS message_client_id,
		m.sender_id AS message_sender_id,
		m.recipient_id AS message_recipient_id,
		m.content AS message_content,
		m.message_type AS message_type,
		m.status AS message_status,
		m.is_delivered AS message_is_delivered,
		m.is_read AS message_is_read,
		m.created_at AS message_created_at,
		m.created_at AS last_activity,
		ROW_NUMBER() OVER (
			PARTITION BY CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END
			ORDER BY m.created_at DESC, m.id DESC
		) AS rn,
		SUM(CASE WHEN m.recipient_id = ? AND m.is_read = false THEN 1 ELSE 0 END) OVER (
			PARTITION BY CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END
		) AS unread_count
	FROM messages m
	WHERE
		m.group_id IS NULL
		AND m.recipient_id IS NOT NULL
		AND (m.sender_id = ? OR m.recipient_id = ?)
)
SELECT
	t.peer_id,
	peer.username AS peer_username,
	peer.email AS peer_email,
	peer.full_name AS peer_full_name,
	peer.avatar AS peer_avatar,
	peer.is_online AS peer_is_online,
	peer.last_seen AS peer_last_seen,
	t.unread_count,
	t.message_id,
	t.message_client_id,
	t.message_sender_id,
	t.message_recipient_id,
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
JOIN users peer ON peer.id = t.peer_id
JOIN users sender ON sender.id = t.message_sender_id
WHERE t.rn = 1
` + "\n" + whereCursor + `
ORDER BY t.last_activity DESC, t.message_id DESC
LIMIT ?
`)

	var rows []ConversationRow
	err := r.db.Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

var _ = gorm.ErrRecordNotFound
