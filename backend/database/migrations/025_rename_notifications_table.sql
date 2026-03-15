-- Migration 043: Ensure the notifications table is named user_notifications.
-- The Notification entity's TableName() returns "user_notifications".
-- If the table was previously created as "notifications", rename it.
-- If it already exists as "user_notifications", this is a no-op.

DO $$
BEGIN
    -- Rename only if old name exists and new name does not
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'notifications'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'user_notifications'
    ) THEN
        ALTER TABLE notifications RENAME TO user_notifications;
    END IF;
END $$;

-- Ensure the table exists with the correct structure even if neither name existed
CREATE TABLE IF NOT EXISTS user_notifications (
    id                    UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id               UUID        REFERENCES users(id) ON DELETE CASCADE,
    template_id           UUID,
    title                 TEXT        NOT NULL,
    body                  TEXT        NOT NULL,
    notification_type     TEXT        NOT NULL,
    reference_id          UUID,
    reference_type        TEXT,
    channels              JSONB,
    is_read               BOOLEAN     NOT NULL DEFAULT FALSE,
    read_at               TIMESTAMPTZ,
    delivery_status       JSONB,
    delivery_attempts     INTEGER     NOT NULL DEFAULT 0,
    last_delivery_attempt TIMESTAMPTZ,
    priority              TEXT        NOT NULL DEFAULT 'normal',
    scheduled_for         TIMESTAMPTZ,
    expires_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON user_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_user_notifications_is_read ON user_notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_user_notifications_type    ON user_notifications(notification_type);
