-- Stock count documents: a multi-line count sheet against one inventory,
-- draft -> posted. Posting compares each line's counted quantity against
-- *live* on-hand (not the snapshot taken at save time) and writes one
-- stock_movements row per SKU whose count differs, via applyStockMovement,
-- so the count carries the same before/after/actor audit as everything
-- else. Saving a draft never touches stock.
CREATE TYPE stock_count_status AS ENUM ('draft', 'posted', 'cancelled');

CREATE TABLE IF NOT EXISTS stock_counts (
    id           bigserial PRIMARY KEY,
    inventory_id bigint NOT NULL REFERENCES inventories(id),
    code         text NOT NULL DEFAULT '',
    status       stock_count_status NOT NULL DEFAULT 'draft',
    note         text NOT NULL DEFAULT '',
    counted_at   timestamp(0) WITH TIME ZONE NOT NULL DEFAULT now(),
    posted_at    timestamp(0) WITH TIME ZONE,
    created_by   bigint REFERENCES users(id),
    posted_by    bigint REFERENCES users(id),
    created_at   timestamp(0) WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS stock_count_items (
    id               bigserial PRIMARY KEY,
    stock_count_id   bigint NOT NULL REFERENCES stock_counts(id) ON DELETE CASCADE,
    variant_id       bigint NOT NULL REFERENCES product_variants(id),
    system_quantity  bigint NOT NULL DEFAULT 0,
    counted_quantity bigint NOT NULL DEFAULT 0,
    difference       bigint NOT NULL DEFAULT 0,
    UNIQUE (stock_count_id, variant_id)
);

CREATE INDEX IF NOT EXISTS idx_stock_counts_inventory ON stock_counts (inventory_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stock_count_items_count ON stock_count_items (stock_count_id);
