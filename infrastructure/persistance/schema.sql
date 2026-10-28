CREATE TABLE IF NOT EXISTS servers (
    id TEXT PRIMARY KEY,
    tickets JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- users/roles/channels intentionally have no FK to servers: a server_id can
-- accrue these rows before a servers row exists for it (EditTickets is the
-- only writer of servers), mirroring the old Firestore subcollection behavior.
CREATE TABLE IF NOT EXISTS users (
    server_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    points DOUBLE PRECISION NOT NULL DEFAULT 0,
    warnings BIGINT NOT NULL DEFAULT 0,
    tickets BIGINT NOT NULL DEFAULT 0,
    flags BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (server_id, user_id)
);

-- Supports both the top-5 leaderboard query and the per-user rank count query.
CREATE INDEX IF NOT EXISTS idx_users_server_points ON users (server_id, points DESC);

CREATE TABLE IF NOT EXISTS user_warnings (
    id BIGSERIAL PRIMARY KEY,
    server_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (server_id, user_id) REFERENCES users (server_id, user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_warnings_user ON user_warnings (server_id, user_id);

CREATE TABLE IF NOT EXISTS roles (
    server_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    flags BIGINT NOT NULL DEFAULT 0,
    level BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (server_id, role_id)
);

CREATE TABLE IF NOT EXISTS channels (
    server_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    flags BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (server_id, channel_id)
);

CREATE TABLE IF NOT EXISTS admins (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);
