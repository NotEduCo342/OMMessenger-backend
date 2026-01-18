-- Create group_read_states for per-member read tracking
CREATE TABLE IF NOT EXISTS group_read_states (
    group_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    last_read_message_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_group_read_states_group_id ON group_read_states (group_id);
CREATE INDEX IF NOT EXISTS idx_group_read_states_user_id ON group_read_states (user_id);

-- Backfill from existing group_members
INSERT INTO group_read_states (group_id, user_id, last_read_message_id, created_at, updated_at)
SELECT gm.group_id, gm.user_id, 0, NOW(), NOW()
FROM group_members gm
ON CONFLICT (group_id, user_id) DO NOTHING;