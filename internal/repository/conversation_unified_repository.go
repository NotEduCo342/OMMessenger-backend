package repository

import (
	"database/sql"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ConversationUnifiedRow is a denormalized row representing either a DM or group conversation
// with last message + unread count + peer/group info.
type ConversationUnifiedRow struct {
	ConversationType string         `gorm:"column:conversation_type"`
	PeerID           sql.NullInt64  `gorm:"column:peer_id"`
	PeerUsername     sql.NullString `gorm:"column:peer_username"`
	PeerEmail        sql.NullString `gorm:"column:peer_email"`
	PeerFullName     sql.NullString `gorm:"column:peer_full_name"`
	PeerAvatar       sql.NullString `gorm:"column:peer_avatar"`
	PeerIsOnline     sql.NullBool   `gorm:"column:peer_is_online"`
	PeerLastSeen     *time.Time     `gorm:"column:peer_last_seen"`

	GroupID            sql.NullInt64  `gorm:"column:group_id"`
	GroupName          sql.NullString `gorm:"column:group_name"`
	GroupIcon          sql.NullString `gorm:"column:group_icon"`
	MemberCount        sql.NullInt64  `gorm:"column:member_count"`
	UnreadCount        int64          `gorm:"column:unread_count"`
	MessageID          uint           `gorm:"column:message_id"`
	MessageClientID    string         `gorm:"column:message_client_id"`
	MessageSenderID    uint           `gorm:"column:message_sender_id"`
	MessageRecipientID sql.NullInt64  `gorm:"column:message_recipient_id"`
	MessageGroupID     sql.NullInt64  `gorm:"column:message_group_id"`
	MessageContent     string         `gorm:"column:message_content"`
	MessageType        string         `gorm:"column:message_type"`
	MessageStatus      string         `gorm:"column:message_status"`
	MessageIsDelivered bool           `gorm:"column:message_is_delivered"`
	MessageIsRead      bool           `gorm:"column:message_is_read"`
	MessageCreatedAt   time.Time      `gorm:"column:message_created_at"`
	LastActivity       time.Time      `gorm:"column:last_activity"`

	SenderID       uint       `gorm:"column:sender_id"`
	SenderUsername string     `gorm:"column:sender_username"`
	SenderEmail    string     `gorm:"column:sender_email"`
	SenderFullName string     `gorm:"column:sender_full_name"`
	SenderAvatar   string     `gorm:"column:sender_avatar"`
	SenderIsOnline bool       `gorm:"column:sender_is_online"`
	SenderLastSeen *time.Time `gorm:"column:sender_last_seen"`
}

func (r *MessageRepository) ListConversationsUnified(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]ConversationUnifiedRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	limitPlusOne := limit + 1

	var whereCursor string
	args := []interface{}{
		userID, userID, userID, userID, userID, userID, userID, // dm_ranked peer/unread
		userID, userID, // group_ranked member join + read state
		userID, // group_empty member join
	}
	if cursorCreatedAt != nil && cursorMessageID > 0 {
		whereCursor = "AND (c.last_activity < ? OR (c.last_activity = ? AND c.message_id < ?))"
		args = append(args, *cursorCreatedAt, *cursorCreatedAt, cursorMessageID)
	}
	args = append(args, limitPlusOne)

	query := strings.TrimSpace(`
WITH dm_ranked AS (
	SELECT
		'dm'::text AS conversation_type,
		CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END AS peer_id,
		peer.username AS peer_username,
		peer.email AS peer_email,
		peer.full_name AS peer_full_name,
		peer.avatar AS peer_avatar,
		peer.is_online AS peer_is_online,
		peer.last_seen AS peer_last_seen,
		NULL::bigint AS group_id,
		NULL::text AS group_name,
		NULL::text AS group_icon,
		NULL::bigint AS member_count,
		SUM(CASE WHEN m.recipient_id = ? AND m.is_read = false THEN 1 ELSE 0 END) OVER (
			PARTITION BY CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END
		) AS unread_count,
		m.id AS message_id,
		m.client_id AS message_client_id,
		m.sender_id AS message_sender_id,
		m.recipient_id AS message_recipient_id,
		NULL::bigint AS message_group_id,
		m.content AS message_content,
		m.message_type AS message_type,
		m.status AS message_status,
		m.is_delivered AS message_is_delivered,
		m.is_read AS message_is_read,
		m.created_at AS message_created_at,
		m.created_at AS last_activity,
		sender.id AS sender_id,
		sender.username AS sender_username,
		sender.email AS sender_email,
		sender.full_name AS sender_full_name,
		sender.avatar AS sender_avatar,
		sender.is_online AS sender_is_online,
		sender.last_seen AS sender_last_seen,
		ROW_NUMBER() OVER (
			PARTITION BY CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END
			ORDER BY m.created_at DESC, m.id DESC
		) AS rn
	FROM messages m
	JOIN users peer ON peer.id = CASE WHEN m.sender_id = ? THEN m.recipient_id ELSE m.sender_id END
	JOIN users sender ON sender.id = m.sender_id
	WHERE
		m.group_id IS NULL
		AND m.recipient_id IS NOT NULL
		AND (m.sender_id = ? OR m.recipient_id = ?)
),
group_ranked AS (
	SELECT
		'group'::text AS conversation_type,
		NULL::bigint AS peer_id,
		NULL::text AS peer_username,
		NULL::text AS peer_email,
		NULL::text AS peer_full_name,
		NULL::text AS peer_avatar,
		NULL::boolean AS peer_is_online,
		NULL::timestamp AS peer_last_seen,
		g.id AS group_id,
		g.name AS group_name,
		g.icon AS group_icon,
		(
			SELECT COUNT(*)
			FROM group_members gm2
			WHERE gm2.group_id = g.id
		) AS member_count,
		SUM(CASE WHEN m.id > COALESCE(grs.last_read_message_id, 0) THEN 1 ELSE 0 END) OVER (
			PARTITION BY m.group_id
		) AS unread_count,
		m.id AS message_id,
		m.client_id AS message_client_id,
		m.sender_id AS message_sender_id,
		NULL::bigint AS message_recipient_id,
		m.group_id AS message_group_id,
		m.content AS message_content,
		m.message_type AS message_type,
		m.status AS message_status,
		m.is_delivered AS message_is_delivered,
		m.is_read AS message_is_read,
		m.created_at AS message_created_at,
		m.created_at AS last_activity,
		sender.id AS sender_id,
		sender.username AS sender_username,
		sender.email AS sender_email,
		sender.full_name AS sender_full_name,
		sender.avatar AS sender_avatar,
		sender.is_online AS sender_is_online,
		sender.last_seen AS sender_last_seen,
		ROW_NUMBER() OVER (
			PARTITION BY m.group_id
			ORDER BY m.created_at DESC, m.id DESC
		) AS rn
	FROM messages m
	JOIN group_members gm ON gm.group_id = m.group_id AND gm.user_id = ?
	JOIN groups g ON g.id = m.group_id
	LEFT JOIN group_read_states grs ON grs.group_id = m.group_id AND grs.user_id = ?
	JOIN users sender ON sender.id = m.sender_id
	WHERE m.group_id IS NOT NULL
),
group_empty AS (
	SELECT
		'group'::text AS conversation_type,
		NULL::bigint AS peer_id,
		NULL::text AS peer_username,
		NULL::text AS peer_email,
		NULL::text AS peer_full_name,
		NULL::text AS peer_avatar,
		NULL::boolean AS peer_is_online,
		NULL::timestamp AS peer_last_seen,
		g.id AS group_id,
		g.name AS group_name,
		g.icon AS group_icon,
		(
			SELECT COUNT(*)
			FROM group_members gm2
			WHERE gm2.group_id = g.id
		) AS member_count,
		0 AS unread_count,
		0::bigint AS message_id,
		''::text AS message_client_id,
		0::bigint AS message_sender_id,
		NULL::bigint AS message_recipient_id,
		NULL::bigint AS message_group_id,
		''::text AS message_content,
		''::text AS message_type,
		''::text AS message_status,
		false AS message_is_delivered,
		false AS message_is_read,
		g.updated_at AS message_created_at,
		g.updated_at AS last_activity,
		0::bigint AS sender_id,
		''::text AS sender_username,
		''::text AS sender_email,
		''::text AS sender_full_name,
		''::text AS sender_avatar,
		false AS sender_is_online,
		NULL::timestamp AS sender_last_seen,
		1 AS rn
	FROM group_members gm
	JOIN groups g ON g.id = gm.group_id
	WHERE gm.user_id = ?
		AND NOT EXISTS (
			SELECT 1
			FROM messages m
			WHERE m.group_id = g.id
		)
),
combined AS (
	SELECT * FROM dm_ranked WHERE rn = 1
	UNION ALL
	SELECT * FROM group_ranked WHERE rn = 1
	UNION ALL
	SELECT * FROM group_empty WHERE rn = 1
)
SELECT * FROM combined c
WHERE 1=1
` + "\n" + whereCursor + `
ORDER BY c.last_activity DESC, c.message_id DESC
LIMIT ?
`)

	var rows []ConversationUnifiedRow
	if err := r.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

var _ = gorm.ErrRecordNotFound
