-- Performance indexes for the conversations list endpoint.
-- The list query does: WHERE conversations.deleted_at IS NULL ORDER BY conversations.created_at [asc|desc] LIMIT n
-- Sessions and message_pairs are fetched by IN (...) lookups, which need indexes on their FK columns.

-- Composite index: covers the soft-delete filter and the ORDER BY in a single index scan.
CREATE INDEX IF NOT EXISTS idx_conversations_deleted_at_created_at
    ON conversations (deleted_at, created_at);

-- Index on session_id for JOIN / preload lookups.
CREATE INDEX IF NOT EXISTS idx_conversations_session_id
    ON conversations (session_id);

-- Soft-delete filter on sessions (used in every Preload("Session") query).
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at
    ON sessions (deleted_at);

-- Source column filter added in 2025_03_17; index enables fast WHERE LOWER(sessions.source) = ?
CREATE INDEX IF NOT EXISTS idx_sessions_source
    ON sessions (source);

-- FK lookup used when loading message_pairs for a single conversation (by-ID endpoint).
CREATE INDEX IF NOT EXISTS idx_message_pairs_conversation_id
    ON message_pairs (conversation_id);

-- Soft-delete filter on message_pairs.
CREATE INDEX IF NOT EXISTS idx_message_pairs_deleted_at
    ON message_pairs (deleted_at);

-- Mobile filter: users.mobile = ?
CREATE INDEX IF NOT EXISTS idx_users_mobile
    ON users (mobile);
