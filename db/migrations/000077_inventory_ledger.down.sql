-- Reverse 000061_inventory_ledger.up.sql.

ALTER TABLE invoice_products DROP COLUMN IF EXISTS unit_cost;
ALTER TABLE invoice_products DROP COLUMN IF EXISTS variant_id;

ALTER TABLE transfer_items DROP CONSTRAINT IF EXISTS transfer_items_transfer_variant_asset_key;
ALTER TABLE transfer_items DROP CONSTRAINT IF EXISTS transfer_items_variant_xor_asset;
ALTER TABLE transfer_items DROP COLUMN IF EXISTS variant_id;
ALTER TABLE transfer_items ADD COLUMN IF NOT EXISTS product_id bigint REFERENCES products(id);
ALTER TABLE transfer_items
    ADD CONSTRAINT transfer_items_check CHECK (
        (product_id IS NOT NULL AND asset_id IS NULL) OR
        (product_id IS NULL AND asset_id IS NOT NULL)
    );
ALTER TABLE transfer_items
    ADD CONSTRAINT transfer_items_transfer_id_product_id_asset_id_key
        UNIQUE (transfer_id, product_id, asset_id);

ALTER TABLE transfers
    DROP COLUMN IF EXISTS code,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS dispatched_at,
    DROP COLUMN IF EXISTS received_at,
    DROP COLUMN IF EXISTS note;

DROP TYPE IF EXISTS transfer_status;

ALTER TABLE purchase_items DROP CONSTRAINT IF EXISTS purchase_items_purchase_variant_asset_key;
ALTER TABLE purchase_items DROP CONSTRAINT IF EXISTS purchase_items_variant_xor_asset;
ALTER TABLE purchase_items DROP COLUMN IF EXISTS variant_id;
ALTER TABLE purchase_items ADD COLUMN IF NOT EXISTS product_id bigint REFERENCES products(id);
ALTER TABLE purchase_items
    ADD CONSTRAINT purchase_items_check CHECK (
        (product_id IS NOT NULL AND asset_id IS NULL) OR
        (product_id IS NULL AND asset_id IS NOT NULL)
    );
ALTER TABLE purchase_items
    ADD CONSTRAINT purchase_items_purchase_id_product_id_asset_id_key
        UNIQUE (purchase_id, product_id, asset_id);

ALTER TABLE purchases
    DROP COLUMN IF EXISTS inventory_id,
    DROP COLUMN IF EXISTS code,
    DROP COLUMN IF EXISTS currency_code,
    DROP COLUMN IF EXISTS subtotal,
    DROP COLUMN IF EXISTS grand_total,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS received_at,
    DROP COLUMN IF EXISTS note;

DROP TYPE IF EXISTS purchase_status;

ALTER TABLE branches DROP COLUMN IF EXISTS inventory_id;

CREATE TABLE IF NOT EXISTS inventories_products (
    product_id   bigint NOT NULL REFERENCES products(id),
    inventory_id bigint NOT NULL REFERENCES inventories(id),
    quantity     bigint NOT NULL,
    PRIMARY KEY (product_id, inventory_id)
);

ALTER TABLE product_variants DROP COLUMN IF EXISTS avg_cost;

DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS inventory_stock;
DROP TYPE IF EXISTS stock_movement_reason;
