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
SELECT * FROM products
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CountProducts :one
SELECT COUNT(*) FROM products;

-- name: UpdateProduct :exec
UPDATE products
  SET name = $2,
  code = $3,
  description = $4
WHERE id = $1;

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
