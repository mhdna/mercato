-- name: CreatePurchase :one
INSERT INTO purchases (
  supplier_id,
  purchased_at
) 
VALUES ( $1, $2 )
RETURNING *;

-- name: GetPurchase :one
SELECT * FROM purchases
WHERE id = $1 LIMIT 1;

-- name: ListPurchases :many
SELECT * FROM purchases
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: AddPurchaseItem :one
INSERT INTO purchase_items (
  purchase_id,
  product_id,
  asset_id,
  quantity,
  unit_price,
  currency_code
) 
VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: AddPurchasedProduct :one
INSERT INTO product_suppliers (
  product_id,
  supplier_id
) 
VALUES ( $1, $2 )
RETURNING *;

-- name: AddPurchasedProductCost :one
INSERT INTO product_supplier_costs (
  product_supplier_id,
  unit_cost,
  currency_code
) 
VALUES ( $1, $2, $3 )
RETURNING *;
-- name: CountPurchases :one
SELECT COUNT(*) FROM purchases;
