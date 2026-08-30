CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS suppliers_phone_trgm_idx
  ON suppliers USING gin (phone gin_trgm_ops);

CREATE INDEX IF NOT EXISTS suppliers_country_trgm_idx
  ON suppliers USING gin (country gin_trgm_ops);

CREATE INDEX IF NOT EXISTS suppliers_address_trgm_idx
  ON suppliers USING gin (address gin_trgm_ops);
