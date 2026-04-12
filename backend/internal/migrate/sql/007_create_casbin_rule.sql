-- +goose Up

CREATE TABLE IF NOT EXISTS casbin_rule (
    id      BIGSERIAL PRIMARY KEY,
    ptype   TEXT NOT NULL,
    v0      TEXT NOT NULL DEFAULT '',
    v1      TEXT NOT NULL DEFAULT '',
    v2      TEXT NOT NULL DEFAULT '',
    v3      TEXT NOT NULL DEFAULT '',
    v4      TEXT NOT NULL DEFAULT '',
    v5      TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_casbin_rule_unique
    ON casbin_rule (ptype, v0, v1, v2, v3, v4, v5);

CREATE INDEX IF NOT EXISTS idx_casbin_rule_ptype ON casbin_rule (ptype);
CREATE INDEX IF NOT EXISTS idx_casbin_rule_domain ON casbin_rule (v1);

COMMENT ON TABLE casbin_rule IS 'Casbin RBAC policy table';
COMMENT ON COLUMN casbin_rule.v0 IS 'Subject or user/role binding source';
COMMENT ON COLUMN casbin_rule.v1 IS 'Tenant/domain';
COMMENT ON COLUMN casbin_rule.v2 IS 'Object or role binding target/domain field';
COMMENT ON COLUMN casbin_rule.v3 IS 'Action';

-- +goose Down
DROP TABLE IF EXISTS casbin_rule;
