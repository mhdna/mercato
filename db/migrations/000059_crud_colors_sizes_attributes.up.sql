-- Track when colors and attribute values were created so the management
-- UI can show it (sizes already have created_at).
ALTER TABLE colors ADD COLUMN created_at timestamptz NOT NULL DEFAULT NOW();
ALTER TABLE attributes_values ADD COLUMN created_at timestamptz NOT NULL DEFAULT NOW();

-- The original sizes table made name / type / "order" each globally unique,
-- which makes it impossible to have more than one size per type
-- (e.g. Shoes:38 and Shoes:39). Scope uniqueness to the type instead.
ALTER TABLE sizes DROP CONSTRAINT IF EXISTS sizes_name_key;
ALTER TABLE sizes DROP CONSTRAINT IF EXISTS sizes_type_key;
ALTER TABLE sizes DROP CONSTRAINT IF EXISTS sizes_order_key;
ALTER TABLE sizes ADD CONSTRAINT sizes_type_name_key UNIQUE (type, name);
