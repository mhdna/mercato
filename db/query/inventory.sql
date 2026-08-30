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

-- name: DeleteInventories :execrows
DELETE FROM inventories
WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: CountInventories :one
SELECT COUNT(*) FROM inventories;

-- ---------------------------------------------------------------------------
-- Per-variant stock: cached on-hand quantity + moving-average cost.
-- ---------------------------------------------------------------------------

-- name: GetInventoryStock :one
SELECT * FROM inventory_stock
WHERE inventory_id = $1 AND variant_id = $2
LIMIT 1;

-- name: AddInventoryStockQuantity :one
-- Apply a signed quantity delta to one SKU's on-hand in one inventory,
-- creating the row if this SKU has never been stocked there. Cost is left
-- untouched (used for sales, returns, transfers, adjustments).
INSERT INTO inventory_stock (inventory_id, variant_id, quantity)
VALUES ($1, $2, sqlc.arg(quantity))
ON CONFLICT (inventory_id, variant_id)
DO UPDATE SET quantity = inventory_stock.quantity + sqlc.arg(quantity),
             updated_at = now()
RETURNING *;

-- name: ReceiveInventoryStock :one
-- Apply an inbound quantity delta and roll the per-location moving-average
-- cost forward. When existing on-hand is <= 0 the average is just reset to
-- the incoming cost (no meaningful prior average to blend).
INSERT INTO inventory_stock (inventory_id, variant_id, quantity, avg_cost)
VALUES ($1, $2, sqlc.arg(quantity), sqlc.arg(unit_cost))
ON CONFLICT (inventory_id, variant_id)
DO UPDATE SET
  avg_cost = CASE
    WHEN inventory_stock.quantity <= 0 THEN sqlc.arg(unit_cost)
    ELSE (inventory_stock.quantity * inventory_stock.avg_cost
          + sqlc.arg(quantity) * sqlc.arg(unit_cost))
         / NULLIF(inventory_stock.quantity + sqlc.arg(quantity), 0)
  END,
  quantity = inventory_stock.quantity + sqlc.arg(quantity),
  updated_at = now()
RETURNING *;

-- name: ListInventoryStock :many
SELECT
  ist.inventory_id,
  ist.variant_id,
  ist.quantity,
  ist.avg_cost,
  ist.updated_at,
  pv.barcode,
  pv.price,
  p.id   AS product_id,
  p.code AS product_code,
  p.name AS product_name,
  COALESCE(c.name, '') AS color_name,
  COALESCE(s.name, '') AS size_name
FROM inventory_stock ist
JOIN product_variants pv ON pv.id = ist.variant_id
JOIN products p ON p.id = pv.product_id
LEFT JOIN colors c ON c.id = pv.color_id
LEFT JOIN sizes  s ON s.id = pv.size_id
WHERE ist.inventory_id = $1
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%'
       OR pv.barcode ILIKE '%' || sqlc.arg(search) || '%')
ORDER BY p.name, s.name
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountInventoryStock :one
SELECT COUNT(*)
FROM inventory_stock ist
JOIN product_variants pv ON pv.id = ist.variant_id
JOIN products p ON p.id = pv.product_id
WHERE ist.inventory_id = $1
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%'
       OR pv.barcode ILIKE '%' || sqlc.arg(search) || '%');

-- ---------------------------------------------------------------------------
-- Stock movement ledger (append-only).
-- ---------------------------------------------------------------------------

-- name: CreateStockMovement :one
INSERT INTO stock_movements (
  inventory_id,
  variant_id,
  quantity,
  reason,
  reference_type,
  reference_id,
  unit_cost,
  note,
  created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListStockMovements :many
SELECT
  sm.id,
  sm.inventory_id,
  sm.variant_id,
  sm.quantity,
  sm.reason,
  sm.reference_type,
  sm.reference_id,
  sm.unit_cost,
  sm.note,
  sm.created_by,
  sm.created_at,
  i.name AS inventory_name,
  pv.barcode,
  p.code AS product_code,
  p.name AS product_name,
  COALESCE(c.name, '') AS color_name,
  COALESCE(s.name, '') AS size_name
FROM stock_movements sm
JOIN inventories i ON i.id = sm.inventory_id
JOIN product_variants pv ON pv.id = sm.variant_id
JOIN products p ON p.id = pv.product_id
LEFT JOIN colors c ON c.id = pv.color_id
LEFT JOIN sizes  s ON s.id = pv.size_id
WHERE (sqlc.narg(inventory_id)::bigint IS NULL OR sm.inventory_id = sqlc.narg(inventory_id))
  AND (sqlc.narg(variant_id)::bigint IS NULL OR sm.variant_id = sqlc.narg(variant_id))
  AND (sqlc.narg(reason)::stock_movement_reason IS NULL OR sm.reason = sqlc.narg(reason))
ORDER BY sm.created_at DESC, sm.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountStockMovements :one
SELECT COUNT(*)
FROM stock_movements sm
WHERE (sqlc.narg(inventory_id)::bigint IS NULL OR sm.inventory_id = sqlc.narg(inventory_id))
  AND (sqlc.narg(variant_id)::bigint IS NULL OR sm.variant_id = sqlc.narg(variant_id))
  AND (sqlc.narg(reason)::stock_movement_reason IS NULL OR sm.reason = sqlc.narg(reason));
