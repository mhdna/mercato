-- name: CreateProduct :one
INSERT INTO products (
  code,
  name,
  description
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
-- Search matches code/name/description (case-insensitive). Every filter is
-- optional: empty search string, NULL narg, or empty attribute_value_ids
-- array all mean "don't filter on this". attribute_value_ids is an AND --
-- the product must carry every selected value.
SELECT p.id, p.code, p.name, p.description, p.is_active, p.created_at
FROM products p
WHERE (sqlc.arg(search)::text = ''
       OR p.code ILIKE '%' || sqlc.arg(search) || '%'
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.description ILIKE '%' || sqlc.arg(search) || '%')
  AND (sqlc.narg(is_active)::bool IS NULL OR p.is_active = sqlc.narg(is_active))
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR p.created_at >= sqlc.narg(created_from))
  AND (sqlc.narg(created_to)::timestamptz IS NULL OR p.created_at < sqlc.narg(created_to))
  AND (sqlc.narg(has_variants)::bool IS NULL
       OR EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id) = sqlc.narg(has_variants))
  AND (cardinality(sqlc.arg(attribute_value_ids)::bigint[]) = 0
       OR (SELECT count(DISTINCT pa.attribute_value_id)
             FROM products_attributes pa
            WHERE pa.product_id = p.id
              AND pa.attribute_value_id = ANY(sqlc.arg(attribute_value_ids)::bigint[]))
           = cardinality(sqlc.arg(attribute_value_ids)::bigint[]))
ORDER BY p.id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountProducts :one
SELECT count(*)
FROM products p
WHERE (sqlc.arg(search)::text = ''
       OR p.code ILIKE '%' || sqlc.arg(search) || '%'
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.description ILIKE '%' || sqlc.arg(search) || '%')
  AND (sqlc.narg(is_active)::bool IS NULL OR p.is_active = sqlc.narg(is_active))
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR p.created_at >= sqlc.narg(created_from))
  AND (sqlc.narg(created_to)::timestamptz IS NULL OR p.created_at < sqlc.narg(created_to))
  AND (sqlc.narg(has_variants)::bool IS NULL
       OR EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id) = sqlc.narg(has_variants))
  AND (cardinality(sqlc.arg(attribute_value_ids)::bigint[]) = 0
       OR (SELECT count(DISTINCT pa.attribute_value_id)
             FROM products_attributes pa
            WHERE pa.product_id = p.id
              AND pa.attribute_value_id = ANY(sqlc.arg(attribute_value_ids)::bigint[]))
           = cardinality(sqlc.arg(attribute_value_ids)::bigint[]));

-- name: UpdateProduct :exec
UPDATE products
  SET name = $2,
  code = $3,
  description = $4,
  is_active = $5
WHERE id = $1;

-- name: UpsertProductAttribute :exec
INSERT INTO products_attributes (attribute_id, product_id, attribute_value_id)
VALUES ($1, $2, $3)
ON CONFLICT (attribute_id, product_id)
DO UPDATE SET attribute_value_id = EXCLUDED.attribute_value_id;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- -- name: CreatePriceList :one
-- INSERT INTO price_lists (
--   name,
--   is_active,
--   is_default,
--   valid_from,
--   valid_to
-- ) VALUES (
--     $1, $2, $3, $4, $5
-- ) RETURNING *;

-- -- name: ListPriceLists :many
-- SELECT * FROM price_lists
-- ORDER BY name
-- LIMIT $1
-- OFFSET $2;

-- -- name: UpdatePriceList :exec
-- UPDATE price_lists
--   SET name = $2,
--   is_active = $3,
--   is_default = $4,
--   valid_from = $5,
--   valid_to = $6
-- WHERE id = $1;

-- -- name: CreateProductPrice :one
-- INSERT INTO price_list_items (
--   price_list_id,
--   product_id,
--   price
-- ) VALUES (
--     $1, $2, $3
-- ) RETURNING *;

-- -- name: UpdateProductPrice :exec
-- UPDATE price_list_items
--   SET price = $3
-- WHERE product_id = $1 AND price_list_id = $2;

-- -- name: CreateDiscountList :one
-- INSERT INTO discount_lists (
--   name,
--   is_active,
--   is_default,
--   valid_from,
--   valid_to
-- ) VALUES (
--     $1, $2, $3, $4, $5
-- ) RETURNING *;

-- -- name: ListDiscountLists :many
-- SELECT * FROM discount_lists
-- ORDER BY name
-- LIMIT $1
-- OFFSET $2;

-- -- name: UpdateDiscountList :exec
-- UPDATE price_lists
--   SET name = $2,
--   is_active = $3,
--   is_default = $4,
--   valid_from = $5,
--   valid_to = $6
-- WHERE id = $1;

-- -- name: UnsetDefaultPriceList :exec
-- UPDATE price_lists
--   SET is_default = false
-- WHERE is_default = true;

-- -- name: CreateProductDiscount :one
-- INSERT INTO discount_list_items (
--   discount_list_id,
--   product_id,
--   discount
-- ) VALUES (
--     $1, $2, $3
-- ) RETURNING *;

-- -- name: UpdateProductDiscount :exec
-- UPDATE discount_list_items
--   SET discount = $3
-- WHERE product_id = $1 AND discount_list_id = $2;

-- name: GetProductByCode :one
SELECT * FROM products WHERE code = $1 LIMIT 1;
