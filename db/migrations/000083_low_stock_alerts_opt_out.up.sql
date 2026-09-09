-- Lets a product opt out of low-stock alerting entirely (e.g. a
-- discontinued or made-to-order item nobody wants an alert for), regardless
-- of on-hand or threshold.
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS low_stock_alerts_enabled bool NOT NULL DEFAULT true;
