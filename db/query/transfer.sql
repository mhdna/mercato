-- name: CreateTransfer :one
INSERT INTO transfers (
  from_inventory_id,
  to_inventory_id,
  type,
  code,
  note
) VALUES ( $1, $2, $3, $4, $5 )
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT
  t.*,
  fi.name AS from_inventory_name,
  ti.name AS to_inventory_name,
  (SELECT COUNT(*) FROM transfer_items x WHERE x.transfer_id = t.id) AS item_count
FROM transfers t
JOIN inventories fi ON fi.id = t.from_inventory_id
JOIN inventories ti ON ti.id = t.to_inventory_id
ORDER BY t.id DESC
LIMIT $1
OFFSET $2;

-- name: UpdateTransfer :exec
UPDATE transfers
SET from_inventory_id = $2,
    to_inventory_id = $3,
    type = $4,
    note = $5
WHERE id = $1 AND status = 'draft';

-- name: SetTransferStatus :one
UPDATE transfers
SET status = sqlc.arg(status),
    dispatched_at = COALESCE(sqlc.narg(dispatched_at), dispatched_at),
    received_at = COALESCE(sqlc.narg(received_at), received_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateTransferItem :one
INSERT INTO transfer_items (
  transfer_id,
  variant_id,
  asset_id,
  quantity
) VALUES ( $1, $2, $3, $4 )
RETURNING *;

-- name: DeleteTransferItem :exec
DELETE FROM transfer_items
WHERE transfer_items.id = $1
  AND transfer_items.transfer_id IN (
    SELECT transfers.id FROM transfers
    WHERE transfers.id = $2 AND transfers.status = 'draft'
  );

-- name: ListTransferItems :many
SELECT
  ti.*,
  p.name AS product_name,
  p.code AS product_code,
  pv.barcode AS variant_barcode,
  a.name AS asset_name
FROM transfer_items ti
LEFT JOIN product_variants pv ON pv.id = ti.variant_id
LEFT JOIN products p ON p.id = pv.product_id
LEFT JOIN assets a ON a.id = ti.asset_id
WHERE ti.transfer_id = $1
ORDER BY ti.id;

-- name: CountTransfers :one
SELECT COUNT(*) FROM transfers;
