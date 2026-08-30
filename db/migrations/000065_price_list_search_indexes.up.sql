CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS price_lists_name_trgm_idx
  ON price_lists USING gin (name gin_trgm_ops);
