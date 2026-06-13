-- Create deleted_conversations table to handle conversation clear/delete operations
CREATE TABLE IF NOT EXISTS deleted_conversations (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  conversation_id VARCHAR(50) NOT NULL,
  cleared_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT idx_user_convo UNIQUE (user_id, conversation_id)
);

CREATE INDEX IF NOT EXISTS idx_deleted_conversations_user_id ON deleted_conversations (user_id);
CREATE INDEX IF NOT EXISTS idx_deleted_conversations_user_convo ON deleted_conversations (user_id, conversation_id);
