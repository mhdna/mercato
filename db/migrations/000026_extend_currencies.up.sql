-- Brings kashi's currencies up to the shape kashi-pos actually needs for
-- real cash-rounding (not just display): is_active lets a currency be
-- retired without deleting history, units_per_usd_micros/cash_rounding_unit
-- are the generalized conversion/rounding fields kashi-pos already added
-- in its own 000052_currency_rates.up.sql. updated_at backs the sync
-- catch-up cursor (Phase 3) — explicit on every write, no trigger, matching
-- how the rest of this codebase handles timestamps.
ALTER TABLE currencies ADD COLUMN is_active boolean not null default true;
ALTER TABLE currencies ADD COLUMN units_per_usd_micros bigint not null default 1000000
  CHECK (units_per_usd_micros > 0);
ALTER TABLE currencies ADD COLUMN cash_rounding_unit bigint not null default 1
  CHECK (cash_rounding_unit > 0);
ALTER TABLE currencies ADD COLUMN updated_at timestamp(0) with time zone not null default now();

UPDATE currencies SET units_per_usd_micros = 100000000, cash_rounding_unit = 1 WHERE code = 'USC';
UPDATE currencies SET units_per_usd_micros = 1000000, cash_rounding_unit = 1 WHERE code = 'USD';
