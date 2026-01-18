-- Add public/private and handle support to groups
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS handle VARCHAR(32);

CREATE UNIQUE INDEX IF NOT EXISTS idx_groups_handle_unique
  ON groups (handle)
  WHERE handle IS NOT NULL;

-- Invite links for private/public groups
CREATE TABLE IF NOT EXISTS group_invite_links (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  token VARCHAR(64) NOT NULL UNIQUE,
  created_by BIGINT NOT NULL REFERENCES users(id),
  expires_at TIMESTAMPTZ,
  max_uses INT,
  used_count INT NOT NULL DEFAULT 0,
  revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_group_invite_links_group_id
  ON group_invite_links (group_id);
