-- name: CreateStockCount :one
INSERT INTO stock_counts (
  inventory_id,
  code,
  note,
  created_by
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetStockCount :one
SELECT * FROM stock_counts
WHERE id = $1 LIMIT 1;

-- name: ListStockCounts :many
SELECT
  sc.*,
  i.name AS inventory_name
FROM stock_counts sc
JOIN inventories i ON i.id = sc.inventory_id
WHERE (sqlc.narg(inventory_id)::bigint IS NULL OR sc.inventory_id = sqlc.narg(inventory_id))
  AND (sqlc.narg(status)::stock_count_status IS NULL OR sc.status = sqlc.narg(status))
ORDER BY sc.created_at DESC, sc.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountStockCounts :one
SELECT COUNT(*)
FROM stock_counts sc
WHERE (sqlc.narg(inventory_id)::bigint IS NULL OR sc.inventory_id = sqlc.narg(inventory_id))
  AND (sqlc.narg(status)::stock_count_status IS NULL OR sc.status = sqlc.narg(status));

-- name: SetStockCountNote :exec
UPDATE stock_counts
SET note = $2
WHERE id = $1 AND status = 'draft';

-- name: SetStockCountStatus :one
UPDATE stock_counts
SET status = $2,
    posted_at = CASE WHEN $2 = 'posted' THEN now() ELSE posted_at END,
    posted_by = CASE WHEN $2 = 'posted' THEN sqlc.arg(posted_by) ELSE posted_by END
WHERE id = $1
RETURNING *;

-- name: DeleteStockCountItems :exec
DELETE FROM stock_count_items
WHERE stock_count_id = $1;

-- name: AddStockCountItem :one
INSERT INTO stock_count_items (
  stock_count_id,
  variant_id,
  system_quantity,
  counted_quantity,
  difference
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListStockCountItems :many
SELECT
  sci.*,
  pv.barcode,
  p.code AS product_code,
  p.name AS product_name,
  COALESCE(c.name, '') AS color_name,
  COALESCE(s.name, '') AS size_name
FROM stock_count_items sci
JOIN product_variants pv ON pv.id = sci.variant_id
JOIN products p ON p.id = pv.product_id
LEFT JOIN colors c ON c.id = pv.color_id
LEFT JOIN sizes  s ON s.id = pv.size_id
WHERE sci.stock_count_id = $1
ORDER BY p.name, s.name;
