CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS purchases_code_trgm_idx
  ON purchases USING gin (code gin_trgm_ops);

CREATE INDEX IF NOT EXISTS purchases_currency_code_trgm_idx
  ON purchases USING gin (currency_code gin_trgm_ops);

CREATE INDEX IF NOT EXISTS suppliers_name_trgm_idx
  ON suppliers USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS inventories_name_trgm_idx
  ON inventories USING gin (name gin_trgm_ops);
