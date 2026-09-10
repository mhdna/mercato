-- Proper inventory: per-variant (SKU) stock, an append-only movement
-- ledger, two-step transfers, purchase invoices that receive stock and
-- drive moving-average cost, and a store inventory per branch so branch
-- (POS) sales can decrement central stock.
--
-- Stock is tracked per product_variants row, not per product. on-hand
-- quantity and moving-average cost live in inventory_stock as a cache;
-- stock_movements is the source of truth (every purchase/sale/return/
-- transfer/adjustment writes one row).

-- ---------------------------------------------------------------------------
-- Stock ledger + cache
-- ---------------------------------------------------------------------------

CREATE TYPE stock_movement_reason AS ENUM (
    'purchase',
    'sale',
    'return',
    'transfer_out',
    'transfer_in',
    'adjustment',
    'count'
);

-- Cached on-hand quantity and per-location moving-average unit cost for
-- one SKU in one inventory. Rebuildable from stock_movements.
CREATE TABLE IF NOT EXISTS inventory_stock (
    inventory_id bigint NOT NULL REFERENCES inventories(id),
    variant_id   bigint NOT NULL REFERENCES product_variants(id),
    quantity     bigint NOT NULL DEFAULT 0,
    avg_cost     bigint NOT NULL DEFAULT 0,
    updated_at   timestamp(0) WITH TIME ZONE NOT NULL DEFAULT now(),
    PRIMARY KEY (inventory_id, variant_id)
);

-- Append-only. quantity is signed (negative = stock leaving). reference_*
-- points at the document that caused the movement (a purchase, invoice,
-- transfer); null for a bare adjustment/count. unit_cost is set for
-- inbound movements that carry a cost (purchase, transfer_in).
CREATE TABLE IF NOT EXISTS stock_movements (
    id             bigserial PRIMARY KEY,
    inventory_id   bigint NOT NULL REFERENCES inventories(id),
    variant_id     bigint NOT NULL REFERENCES product_variants(id),
    quantity       bigint NOT NULL,
    reason         stock_movement_reason NOT NULL,
    reference_type text,
    reference_id   bigint,
    unit_cost      bigint,
    note           text NOT NULL DEFAULT '',
    created_by     bigint REFERENCES users(id),
    created_at     timestamp(0) WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_inv_variant
    ON stock_movements (inventory_id, variant_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_reference
    ON stock_movements (reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at
    ON stock_movements (created_at DESC);

-- Global moving-average cost across every location, maintained on purchase
-- receive. Sales lines snapshot this as COGS.
ALTER TABLE product_variants
    ADD COLUMN IF NOT EXISTS avg_cost bigint NOT NULL DEFAULT 0;

-- inventories_products was a product-level running counter; inventory_stock
-- (variant-level, with cost, backed by the ledger) replaces it.
DROP TABLE IF EXISTS inventories_products;

-- ---------------------------------------------------------------------------
-- One store inventory per branch
-- ---------------------------------------------------------------------------

ALTER TABLE branches
    ADD COLUMN IF NOT EXISTS inventory_id bigint REFERENCES inventories(id);

-- Backfill: give every existing branch its own store inventory.
INSERT INTO inventories (name, type, code)
SELECT 'Branch: ' || b.name, 'store', 'BRANCH-' || b.code
FROM branches b
WHERE b.inventory_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM inventories i WHERE i.code = 'BRANCH-' || b.code);

UPDATE branches b
SET inventory_id = i.id
FROM inventories i
WHERE i.code = 'BRANCH-' || b.code
  AND b.inventory_id IS NULL;

-- ---------------------------------------------------------------------------
-- Purchase invoices
-- ---------------------------------------------------------------------------

CREATE TYPE purchase_status AS ENUM ('draft', 'received', 'cancelled');

ALTER TABLE purchases
    ADD COLUMN IF NOT EXISTS inventory_id  bigint REFERENCES inventories(id),
    ADD COLUMN IF NOT EXISTS code          text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS currency_code text NOT NULL DEFAULT 'USD' REFERENCES currencies(code),
    ADD COLUMN IF NOT EXISTS subtotal      bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS grand_total   bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS status        purchase_status NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS received_at   timestamp(0) WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS note          text NOT NULL DEFAULT '';

-- Trigram search indexes for the columns just added. Migration 63 wanted
-- these but ran before the columns existed, so it skips them and they land
-- here instead.
CREATE INDEX IF NOT EXISTS purchases_code_trgm_idx
    ON purchases USING gin (code gin_trgm_ops);
CREATE INDEX IF NOT EXISTS purchases_currency_code_trgm_idx
    ON purchases USING gin (currency_code gin_trgm_ops);

-- purchase_items move from product to variant. asset purchases stay.
-- Pre-launch: the old product-keyed rows never drove stock and can't be
-- mapped 1:1 to a variant, so they are cleared rather than migrated.
DELETE FROM purchase_items;
ALTER TABLE purchase_items DROP CONSTRAINT IF EXISTS purchase_items_check;
ALTER TABLE purchase_items DROP CONSTRAINT IF EXISTS purchase_items_purchase_id_product_id_asset_id_key;
ALTER TABLE purchase_items DROP COLUMN IF EXISTS product_id;
ALTER TABLE purchase_items
    ADD COLUMN IF NOT EXISTS variant_id bigint REFERENCES product_variants(id);
ALTER TABLE purchase_items
    ADD CONSTRAINT purchase_items_variant_xor_asset CHECK (
        (variant_id IS NOT NULL AND asset_id IS NULL) OR
        (variant_id IS NULL AND asset_id IS NOT NULL)
    );
ALTER TABLE purchase_items
    ADD CONSTRAINT purchase_items_purchase_variant_asset_key
        UNIQUE (purchase_id, variant_id, asset_id);

-- ---------------------------------------------------------------------------
-- Two-step transfers
-- ---------------------------------------------------------------------------

CREATE TYPE transfer_status AS ENUM ('draft', 'dispatched', 'received', 'cancelled');

ALTER TABLE transfers
    ADD COLUMN IF NOT EXISTS code          text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status        transfer_status NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS dispatched_at timestamp(0) WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS received_at   timestamp(0) WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS note          text NOT NULL DEFAULT '';

DELETE FROM transfer_items;
ALTER TABLE transfer_items DROP CONSTRAINT IF EXISTS transfer_items_check;
ALTER TABLE transfer_items DROP CONSTRAINT IF EXISTS transfer_items_transfer_id_product_id_asset_id_key;
ALTER TABLE transfer_items DROP COLUMN IF EXISTS product_id;
ALTER TABLE transfer_items
    ADD COLUMN IF NOT EXISTS variant_id bigint REFERENCES product_variants(id);
ALTER TABLE transfer_items
    ADD CONSTRAINT transfer_items_variant_xor_asset CHECK (
        (variant_id IS NOT NULL AND asset_id IS NULL) OR
        (variant_id IS NULL AND asset_id IS NOT NULL)
    );
ALTER TABLE transfer_items
    ADD CONSTRAINT transfer_items_transfer_variant_asset_key
        UNIQUE (transfer_id, variant_id, asset_id);

-- ---------------------------------------------------------------------------
-- Sales invoice lines reference a specific SKU and snapshot COGS
-- ---------------------------------------------------------------------------

ALTER TABLE invoice_products
    ADD COLUMN IF NOT EXISTS variant_id bigint REFERENCES product_variants(id),
    ADD COLUMN IF NOT EXISTS unit_cost  bigint NOT NULL DEFAULT 0;

-- Best-effort backfill for any existing rows: first variant of the product.
UPDATE invoice_products ip
SET variant_id = (
    SELECT pv.id FROM product_variants pv
    WHERE pv.product_id = ip.product_id
    ORDER BY pv.id
    LIMIT 1
)
WHERE ip.variant_id IS NULL;
