-- +goose Up

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_rule_source enum
-- auto = derived by Reminder Worker from OCR
-- manual = user-created
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE reminder_rule_source AS ENUM ('auto', 'manual');

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_rules
-- Rules that define when to send reminders for a document.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE reminder_rules (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id           UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id             UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rule_type             TEXT NOT NULL,              -- expiry, renewal, due_date
    trigger_date          DATE NOT NULL,              -- the date that triggers the reminder
    notify_days_before    INTEGER[] NOT NULL DEFAULT '{30, 7, 1, 0}',  -- days before trigger_date to notify
    source                reminder_rule_source NOT NULL DEFAULT 'auto',
    active                BOOLEAN NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reminder_rules_document ON reminder_rules(document_id);
CREATE INDEX idx_reminder_rules_tenant ON reminder_rules(tenant_id);
CREATE INDEX idx_reminder_rules_trigger ON reminder_rules(trigger_date);
CREATE INDEX idx_reminder_rules_active ON reminder_rules(active) WHERE active = true;

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_event_status enum
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TYPE reminder_event_status AS ENUM ('pending', 'sent', 'failed', 'snoozed');

-- ═══════════════════════════════════════════════════════════════════════════════
-- reminder_events
-- Tracks delivery state per notification.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE reminder_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id         UUID NOT NULL REFERENCES reminder_rules(id) ON DELETE CASCADE,
    scheduled_at    TIMESTAMPTZ NOT NULL,              -- when the notification should be sent
    sent_at         TIMESTAMPTZ,                       -- when it was actually sent
    channel         TEXT NOT NULL DEFAULT 'in_app',    -- in_app, email
    status          reminder_event_status NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reminder_events_rule ON reminder_events(rule_id);
CREATE INDEX idx_reminder_events_status ON reminder_events(status);
CREATE INDEX idx_reminder_events_scheduled ON reminder_events(scheduled_at) WHERE status = 'pending';

-- ═══════════════════════════════════════════════════════════════════════════════
-- audit_events
-- Append-only audit log. Cannot be modified by any user role.
-- ═══════════════════════════════════════════════════════════════════════════════
CREATE TABLE audit_events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    actor_id    UUID REFERENCES users(id),             -- null for system actions
    entity_type TEXT NOT NULL,                          -- document, user, folder, reminder
    entity_id   UUID NOT NULL,
    action      TEXT NOT NULL,                          -- upload, download, delete, metadata.correct, member.invite, etc.
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_tenant ON audit_events(tenant_id);
CREATE INDEX idx_audit_actor ON audit_events(actor_id);
CREATE INDEX idx_audit_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_action ON audit_events(action);
CREATE INDEX idx_audit_created ON audit_events(created_at DESC);

-- Prevent updates and deletes on audit_events (enforce append-only)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only; updates and deletes are not permitted';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER audit_events_no_update
    BEFORE UPDATE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();

CREATE TRIGGER audit_events_no_delete
    BEFORE DELETE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();

-- +goose Down
DROP TRIGGER IF EXISTS audit_events_no_delete ON audit_events;
DROP TRIGGER IF EXISTS audit_events_no_update ON audit_events;
DROP FUNCTION IF EXISTS prevent_audit_modification();
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS reminder_events;
DROP TABLE IF EXISTS reminder_rules;
DROP TYPE IF EXISTS reminder_event_status;
DROP TYPE IF EXISTS reminder_rule_source;
