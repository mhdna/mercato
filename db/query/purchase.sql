-- name: CreatePurchase :one
INSERT INTO purchases (
  supplier_id,
  inventory_id,
  code,
  currency_code,
  purchased_at,
  note
)
VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: GetPurchase :one
SELECT
  p.*,
  s.name AS supplier_name,
  i.name AS inventory_name
FROM purchases p
JOIN suppliers s ON s.id = p.supplier_id
LEFT JOIN inventories i ON i.id = p.inventory_id
WHERE p.id = $1 LIMIT 1;

-- name: ListPurchases :many
SELECT
  p.*,
  s.name AS supplier_name,
  i.name AS inventory_name,
  (SELECT COUNT(*) FROM purchase_items x WHERE x.purchase_id = p.id) AS item_count
FROM purchases p
JOIN suppliers s ON s.id = p.supplier_id
LEFT JOIN inventories i ON i.id = p.inventory_id
WHERE (
  BTRIM(sqlc.arg(search)::text) = ''
  OR p.code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR s.name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR COALESCE(i.name, '') ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR p.status::text ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR p.currency_code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
)
ORDER BY p.id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: SetPurchaseTotals :exec
UPDATE purchases
SET subtotal = $2,
    grand_total = $3
WHERE id = $1;

-- name: SetPurchaseStatus :one
UPDATE purchases
SET status = sqlc.arg(status),
    received_at = COALESCE(sqlc.narg(received_at), received_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AddPurchaseItem :one
INSERT INTO purchase_items (
  purchase_id,
  variant_id,
  asset_id,
  quantity,
  unit_price,
  currency_code
)
VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: DeletePurchaseItem :exec
DELETE FROM purchase_items
WHERE purchase_items.id = $1
  AND purchase_items.purchase_id IN (
    SELECT purchases.id FROM purchases
    WHERE purchases.id = $2 AND purchases.status = 'draft'
  );

-- name: ListPurchaseItems :many
SELECT
  pi.*,
  p.name AS product_name,
  p.code AS product_code,
  pv.barcode AS variant_barcode,
  a.name AS asset_name
FROM purchase_items pi
LEFT JOIN product_variants pv ON pv.id = pi.variant_id
LEFT JOIN products p ON p.id = pv.product_id
LEFT JOIN assets a ON a.id = pi.asset_id
WHERE pi.purchase_id = $1
ORDER BY pi.id;

-- name: GetProductSupplier :one
SELECT * FROM product_suppliers
WHERE product_id = $1 AND supplier_id = $2
LIMIT 1;

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
ON CONFLICT (product_supplier_id, unit_cost) DO NOTHING
RETURNING *;

-- name: CountPurchasesFiltered :one
SELECT COUNT(*)
FROM purchases p
JOIN suppliers s ON s.id = p.supplier_id
LEFT JOIN inventories i ON i.id = p.inventory_id
WHERE (
  BTRIM(sqlc.arg(search)::text) = ''
  OR p.code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR s.name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR COALESCE(i.name, '') ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR p.status::text ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR p.currency_code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
);
