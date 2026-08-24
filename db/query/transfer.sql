-- name: CreateTransfer :one
INSERT INTO transfers (
  from_inventory_id,
  to_inventory_id,
  type
) VALUES ( $1, $2, $3 )
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateTransfer :exec
UPDATE transfers 
SET from_inventory_id = $2,
to_inventory_id = $3,
type = $4
WHERE id = $1;

-- name: CreateTransferItem :one
INSERT INTO transfer_items (
  transfer_id,
  product_id,
  asset_id,
  quantity
) VALUES ( $1, $2, $3, $4 )
RETURNING *;

-- TODO maybe this is not so clean
-- name: ListTransferItems :many
SELECT t.*, p.name as product_name, a.name as asset_name
FROM transfer_items t
left join products p on product_id = p.id
left join assets a on asset_id = a.id
where t.transfer_id = $1;

-- name: CountTransfers :one
SELECT COUNT(*) FROM transfers;
