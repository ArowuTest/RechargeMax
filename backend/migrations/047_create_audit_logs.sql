-- Audit log table for recording all state changes in the system
CREATE TABLE IF NOT EXISTS audit_logs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_user_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    action        VARCHAR(100) NOT NULL,          -- e.g. USER_BANNED, SETTING_UPDATED
    entity_type   VARCHAR(100),                   -- e.g. user, platform_setting, affiliate
    entity_id     TEXT,                           -- the affected row's ID
    old_value     JSONB,                          -- before state (null for creates)
    new_value     JSONB,                          -- after state (null for deletes)
    ip_address    INET,
    user_agent    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_admin_user_id  ON audit_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id        ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action         ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity         ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at     ON audit_logs(created_at DESC);
