-- name: CreateProductVariant :one
INSERT INTO product_variants (
  product_id,
  color_id,
  size_id,
  barcode,
  price
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetProductVariant :one
SELECT * FROM product_variants
WHERE id = $1 LIMIT 1;

-- name: GetProductVariantByBarcode :one
SELECT * FROM product_variants
WHERE barcode = $1 LIMIT 1;

-- name: ListProductVariantsByProduct :many
SELECT * FROM product_variants
WHERE product_id = $1
ORDER BY id;

-- name: ListProductVariants :many
SELECT sqlc.embed(product_variants), sqlc.embed(products) FROM product_variants
INNER JOIN products ON products.id = product_variants.product_id
ORDER BY product_variants.id
LIMIT $1
OFFSET $2;

-- name: CountProductVariants :one
SELECT COUNT(*) FROM product_variants;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET color_id = $2,
size_id = $3,
barcode = $4,
price = $5,
is_active = $6,
updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListProductVariantsUpdatedSince :many
SELECT * FROM product_variants
WHERE updated_at > $1
ORDER BY updated_at;

-- name: ListProductVariantsForSync :many
-- Flattened for kashi-pos, which has no separate "product" row -- each
-- variant carries everything a branch needs to create or update its own
-- flat product row. Colors/sizes/attributes are resolved to their name
-- here (not sent as ids) since kashi-pos matches everything by name, not
-- by kashi's internal ids.
--
-- Attribute ids (2=brand, 1=sub-category, 3=kind, 7=season, 6=year) match
-- the fixed seed in 000005_create_products_attributes_table.up.sql.
-- Editing a product's attributes doesn't currently bump
-- product_variants.updated_at (no endpoint does that edit yet -- see
-- api/product.go), so an attribute-only change wouldn't be picked up by
-- this filter if one ever existed; today attributes are only ever set at
-- creation time, in the same moment the variant's own updated_at is set,
-- so this doesn't miss anything in practice yet.
SELECT
  product_variants.id,
  product_variants.barcode,
  product_variants.price,
  product_variants.is_active,
  product_variants.updated_at,
  products.code,
  products.name,
  products.description,
  colors.name AS color_name,
  sizes.name AS size_name,
  brand_val.value AS brand,
  subcategory_val.value AS sub_category,
  kind_val.value AS kind,
  season_val.value AS season,
  year_val.value AS year
FROM product_variants
JOIN products ON products.id = product_variants.product_id
LEFT JOIN colors ON colors.id = product_variants.color_id
LEFT JOIN sizes ON sizes.id = product_variants.size_id
LEFT JOIN products_attributes brand_pa ON brand_pa.product_id = products.id AND brand_pa.attribute_id = 2
LEFT JOIN attributes_values brand_val ON brand_val.id = brand_pa.attribute_value_id
LEFT JOIN products_attributes subcategory_pa ON subcategory_pa.product_id = products.id AND subcategory_pa.attribute_id = 1
LEFT JOIN attributes_values subcategory_val ON subcategory_val.id = subcategory_pa.attribute_value_id
LEFT JOIN products_attributes kind_pa ON kind_pa.product_id = products.id AND kind_pa.attribute_id = 3
LEFT JOIN attributes_values kind_val ON kind_val.id = kind_pa.attribute_value_id
LEFT JOIN products_attributes season_pa ON season_pa.product_id = products.id AND season_pa.attribute_id = 7
LEFT JOIN attributes_values season_val ON season_val.id = season_pa.attribute_value_id
LEFT JOIN products_attributes year_pa ON year_pa.product_id = products.id AND year_pa.attribute_id = 6
LEFT JOIN attributes_values year_val ON year_val.id = year_pa.attribute_value_id
WHERE product_variants.updated_at > $1
ORDER BY product_variants.updated_at;
