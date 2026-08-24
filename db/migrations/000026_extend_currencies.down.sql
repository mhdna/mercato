ALTER TABLE currencies DROP COLUMN IF EXISTS updated_at;
ALTER TABLE currencies DROP COLUMN IF EXISTS cash_rounding_unit;
ALTER TABLE currencies DROP COLUMN IF EXISTS units_per_usd_micros;
ALTER TABLE currencies DROP COLUMN IF EXISTS is_active;
