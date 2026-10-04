-- Platform infrastructure tables.
-- audit_logs is written by the AuditLog middleware on every state-mutating request.

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(512) NOT NULL,
    status INT NOT NULL,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);

-- user_connections backs GraphRepository.GetUserNetwork, which walks the
-- friendship graph to a bounded depth via a recursive CTE.
CREATE TABLE IF NOT EXISTS user_connections (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, friend_id),
    CONSTRAINT user_connections_no_self CHECK (user_id <> friend_id)
);

CREATE INDEX IF NOT EXISTS idx_user_connections_friend_id ON user_connections (friend_id);