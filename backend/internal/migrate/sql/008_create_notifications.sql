-- +goose Up

-- ═══════════════════════════════════════════════════════════════════════════════
-- notifications
-- In-app notification records for users.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT NOT NULL DEFAULT 'reminder',
    title       TEXT NOT NULL,
    body        TEXT,
    link        TEXT,
    status      TEXT NOT NULL DEFAULT 'unread',
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at     TIMESTAMPTZ
);

CREATE INDEX idx_notifications_tenant_user ON notifications (tenant_id, user_id);
CREATE INDEX idx_notifications_status ON notifications (status);
CREATE INDEX idx_notifications_created ON notifications (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS notifications;
