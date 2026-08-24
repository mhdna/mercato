-- name: CreateInventory :one
INSERT INTO inventories (
  name,
  type,
  code,
  longitude,
  latitude
) VALUES ( $1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInventory :one
SELECT * FROM inventories
WHERE id = $1 LIMIT 1;

-- name: ListInventories :many
SELECT * FROM inventories
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: UpdateInventory :exec
UPDATE inventories 
SET name = $2
WHERE id = $1;

-- name: DeleteInventory :exec
DELETE FROM inventories
WHERE id = $1;

-- name: AddInventoryProduct :one
INSERT INTO inventories_products (
  inventory_id,
  product_id,
  quantity
)
VALUES ( $1, $2, $3)
RETURNING *;

-- name: ListInventoryProducts :many
SELECT
  ip.inventory_id,
  ip.product_id,
  ip.quantity,
  p.name, 
  p.code
FROM inventories_products ip
JOIN products p ON p.id = ip.product_id
WHERE ip.inventory_id = $1
ORDER BY p.name;

-- name: AddInventoryProductQuantity :exec
UPDATE inventories_products
SET quantity = quantity + $3
WHERE inventory_id = $1
AND product_id = $2;

-- name: DeleteInventoryProduct :exec
DELETE FROM inventories_products
WHERE inventory_id = $1
AND product_id = $2;

-- name: CountInventories :one
SELECT COUNT(*) FROM inventories;
