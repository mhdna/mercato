DROP INDEX IF EXISTS idx_stock_movements_created_by;

ALTER TABLE stock_movements
    DROP COLUMN IF EXISTS quantity_before,
    DROP COLUMN IF EXISTS quantity_after,
    DROP COLUMN IF EXISTS metadata;
