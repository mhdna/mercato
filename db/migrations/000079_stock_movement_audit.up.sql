-- Widen the stock movement ledger into a full audit trail: capture the
-- on-hand quantity immediately before/after each movement, plus an
-- event-specific metadata blob (mirrors kashi-pos-adjacent tooling's
-- inventory_audit table, folded into the existing ledger instead of a
-- parallel table -- stock_movements already carries reason/reference/actor).
--
-- Backfill is not meaningful for existing rows: before/after were never
-- recorded, so they are left at their zero defaults rather than guessed.
ALTER TABLE stock_movements
    ADD COLUMN IF NOT EXISTS quantity_before bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS quantity_after  bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS metadata        jsonb  NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_stock_movements_created_by ON stock_movements (created_by);
