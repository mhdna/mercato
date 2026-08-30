ALTER TABLE sizes DROP CONSTRAINT IF EXISTS sizes_type_name_key;
ALTER TABLE sizes ADD CONSTRAINT sizes_order_key UNIQUE ("order");
ALTER TABLE sizes ADD CONSTRAINT sizes_type_key UNIQUE (type);
ALTER TABLE sizes ADD CONSTRAINT sizes_name_key UNIQUE (name);

ALTER TABLE attributes_values DROP COLUMN created_at;
ALTER TABLE colors DROP COLUMN created_at;
