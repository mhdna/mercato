-- Generic audit log for everything that isn't a stock movement (which has
-- its own richer ledger in stock_movements/inventory_stock). One row per
-- create/update/delete on an audited entity, with a before/after JSON
-- snapshot and the acting user.
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete');

CREATE TABLE IF NOT EXISTS audit_logs (
    id          bigserial PRIMARY KEY,
    entity_type text NOT NULL,
    entity_id   bigint NOT NULL,
    action      audit_action NOT NULL,
    before      jsonb,
    after       jsonb,
    actor_id    bigint REFERENCES users(id),
    source      text NOT NULL DEFAULT 'ui',
    created_at  timestamp(0) WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs (entity_type, entity_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);
