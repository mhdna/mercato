-- name: CreatePriceListItem :one
INSERT INTO price_list_items (price_list_id, product_id, price)
VALUES ($1, $2, $3)
ON CONFLICT (price_list_id, product_id)
DO UPDATE SET price = $3
RETURNING *;

-- name: GetProductPriceFromList :one
SELECT * FROM price_list_items
WHERE price_list_id = $1 AND product_id = $2;

-- name: GetDefaultPriceForProduct :one
SELECT pli.price FROM price_list_items pli
JOIN price_lists pl ON pl.id = pli.price_list_id
WHERE pli.product_id = $1
AND pl.is_active = true AND pl.is_default = true
AND now() >= pl.valid_from AND now() <= pl.valid_to
LIMIT 1;

-- name: UpdatePriceListItem :exec
UPDATE price_list_items
SET price = $3
WHERE price_list_id = $1 AND product_id = $2;

-- name: DeletePriceListItem :exec
DELETE FROM price_list_items
WHERE price_list_id = $1 AND product_id = $2;

-- name: ListPriceListItems :many
SELECT * FROM price_list_items
WHERE price_list_id = $1
ORDER BY product_id;

-- name: ListPriceListItemsWithProduct :many
SELECT pli.price_list_id, pli.product_id, pli.price,
       p.code AS product_code, p.name AS product_name
FROM price_list_items pli
JOIN products p ON p.id = pli.product_id
WHERE pli.price_list_id = $1
ORDER BY p.name;
