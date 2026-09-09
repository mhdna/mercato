-- Low stock alerts: a per-product override threshold, falling back to one
-- global default. "Low stock" means a product's total on-hand (summed
-- across every SKU and every inventory) is at or below its threshold;
-- zero or negative on-hand is "out of stock", a stricter subset of that.
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS low_stock_threshold bigint;

ALTER TABLE app_settings
    ADD COLUMN IF NOT EXISTS default_low_stock_threshold bigint NOT NULL DEFAULT 5;
