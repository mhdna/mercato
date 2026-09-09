ALTER TABLE app_settings
    DROP COLUMN IF EXISTS default_low_stock_threshold;

ALTER TABLE products
    DROP COLUMN IF EXISTS low_stock_threshold;
