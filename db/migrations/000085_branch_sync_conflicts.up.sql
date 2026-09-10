-- Rows a branch pushed that kashi can't safely apply on its own and that an
-- admin must resolve: a client whose phone matches an existing central
-- client under a different name, a client with no usable phone, or a
-- product barcode the till has that kashi has never seen. Staging them here
-- (rather than guessing, or 500ing and retrying forever) is what lets the
-- branch keep syncing everything else while a human sorts these out.
CREATE TABLE IF NOT EXISTS branch_sync_conflicts (
    id bigserial PRIMARY KEY,
    branch_id bigint NOT NULL REFERENCES branches (id),
    entity text NOT NULL,                 -- 'client' | 'product'
    ref text NOT NULL,                    -- branch_client_id (as text) | barcode
    kind text NOT NULL,                   -- 'phone_name_mismatch' | 'invalid_phone' | 'unknown_product'
    branch_payload jsonb NOT NULL,        -- exactly what the till reported
    central_client_id bigint REFERENCES clients (id),
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'dismissed')),
    resolution jsonb,
    resolved_by bigint REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    -- One open row per (branch, entity, ref): a retried outbox push
    -- re-reports the same conflict, and that must update the existing row,
    -- not pile up duplicates.
    UNIQUE (branch_id, entity, ref)
);

CREATE INDEX IF NOT EXISTS idx_branch_sync_conflicts_open
    ON branch_sync_conflicts (branch_id, status);
