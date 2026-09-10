CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- purchases.code / purchases.currency_code are added in migration 77; this
-- migration predates them, so guard the purchase indexes and let 77 create
-- them once the columns exist. suppliers.name / inventories.name are here
-- already, so those are unconditional.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name = 'purchases' AND column_name = 'code') THEN
    CREATE INDEX IF NOT EXISTS purchases_code_trgm_idx
      ON purchases USING gin (code gin_trgm_ops);
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_name = 'purchases' AND column_name = 'currency_code') THEN
    CREATE INDEX IF NOT EXISTS purchases_currency_code_trgm_idx
      ON purchases USING gin (currency_code gin_trgm_ops);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS suppliers_name_trgm_idx
  ON suppliers USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS inventories_name_trgm_idx
  ON inventories USING gin (name gin_trgm_ops);
