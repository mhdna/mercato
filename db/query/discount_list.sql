-- name: CreateDiscountList :one
INSERT INTO discount_lists (name, is_active, is_default, valid_from, valid_to)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDiscountList :one
SELECT * FROM discount_lists
WHERE id = $1 LIMIT 1;

-- name: ListDiscountLists :many
SELECT * FROM discount_lists
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: UpdateDiscountList :exec
UPDATE discount_lists
SET name = $2, is_active = $3, is_default = $4, valid_from = $5, valid_to = $6
WHERE id = $1;

-- name: UnsetDefaultDiscountList :exec
UPDATE discount_lists
SET is_default = false
WHERE is_default = true AND id != $1;

-- name: CreateDiscountListItem :one
INSERT INTO discount_list_items (discount_list_id, product_id, discount)
VALUES ($1, $2, $3)
ON CONFLICT (discount_list_id, product_id)
DO UPDATE SET discount = $3
RETURNING *;

-- name: GetProductDiscountFromList :one
SELECT * FROM discount_list_items
WHERE discount_list_id = $1 AND product_id = $2;

-- name: GetDefaultDiscountForProduct :one
SELECT dli.discount FROM discount_list_items dli
JOIN discount_lists dl ON dl.id = dli.discount_list_id
WHERE dli.product_id = $1
AND dl.is_active = true AND dl.is_default = true
AND now() >= dl.valid_from AND now() <= dl.valid_to
LIMIT 1;

-- name: UpdateDiscountListItem :exec
UPDATE discount_list_items
SET discount = $3
WHERE discount_list_id = $1 AND product_id = $2;

-- name: DeleteDiscountListItem :exec
DELETE FROM discount_list_items
WHERE discount_list_id = $1 AND product_id = $2;

-- name: ListDiscountListItems :many
SELECT * FROM discount_list_items
WHERE discount_list_id = $1
ORDER BY product_id;

-- name: CountDiscountLists :one
SELECT COUNT(*) FROM discount_lists;
