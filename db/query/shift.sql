-- name: CreateShift :one
INSERT INTO shifts (
  cashbox_id
) 
VALUES ( $1 )
RETURNING *;

-- name: GetShift :one
SELECT * FROM shifts
WHERE id = $1
LIMIT 1;

-- name: CloseShift :exec
UPDATE shifts 
  SET closed_at = NOW(),
  is_closed = true
WHERE id = $1;


-- name: ListShifts :many
SELECT
  s.id, s.is_closed, s.cashbox_id, s.created_at, s.closed_at,
  c.name AS cashbox_name,
  c.code AS cashbox_code
FROM shifts s
INNER JOIN cashboxes c ON c.id = s.cashbox_id
WHERE (
    sqlc.arg(status)::text = ''
    OR (sqlc.arg(status)::text = 'open' AND s.is_closed = false)
    OR (sqlc.arg(status)::text = 'closed' AND s.is_closed = true)
  )
  AND (sqlc.narg(cashbox_id)::bigint IS NULL OR s.cashbox_id = sqlc.narg(cashbox_id))
  AND (
    sqlc.arg(search)::text = ''
    OR c.name ILIKE '%' || sqlc.arg(search)::text || '%'
    OR c.code ILIKE '%' || sqlc.arg(search)::text || '%'
    OR CAST(s.id AS text) = sqlc.arg(search)::text
  )
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'cashbox' AND sqlc.arg(sort_order)::text = 'asc' THEN c.name END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'cashbox' AND sqlc.arg(sort_order)::text = 'desc' THEN c.name END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'created_at' AND sqlc.arg(sort_order)::text = 'asc' THEN s.created_at END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'created_at' AND sqlc.arg(sort_order)::text = 'desc' THEN s.created_at END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'closed_at' AND sqlc.arg(sort_order)::text = 'asc' THEN s.closed_at END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'closed_at' AND sqlc.arg(sort_order)::text = 'desc' THEN s.closed_at END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'asc' THEN s.id END ASC,
  s.id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountShifts :one
SELECT COUNT(*)
FROM shifts s
INNER JOIN cashboxes c ON c.id = s.cashbox_id
WHERE (
    sqlc.arg(status)::text = ''
    OR (sqlc.arg(status)::text = 'open' AND s.is_closed = false)
    OR (sqlc.arg(status)::text = 'closed' AND s.is_closed = true)
  )
  AND (sqlc.narg(cashbox_id)::bigint IS NULL OR s.cashbox_id = sqlc.narg(cashbox_id))
  AND (
    sqlc.arg(search)::text = ''
    OR c.name ILIKE '%' || sqlc.arg(search)::text || '%'
    OR c.code ILIKE '%' || sqlc.arg(search)::text || '%'
    OR CAST(s.id AS text) = sqlc.arg(search)::text
  );
