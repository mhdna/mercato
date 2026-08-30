-- name: CreateSize :one
INSERT INTO sizes (
  name,
  type,
  "order"
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: ListSizes :many
SELECT * FROM sizes
ORDER BY type, "order";

-- name: ListSizesPage :many
SELECT * FROM sizes
WHERE (
  sqlc.arg(search)::text = ''
  OR name ILIKE '%' || sqlc.arg(search)::text || '%'
  OR type ILIKE '%' || sqlc.arg(search)::text || '%'
  OR "order" ILIKE '%' || sqlc.arg(search)::text || '%'
)
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'type' AND sqlc.arg(sort_order)::text = 'asc' THEN type END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'type' AND sqlc.arg(sort_order)::text = 'desc' THEN type END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'name' AND sqlc.arg(sort_order)::text = 'asc' THEN name END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'name' AND sqlc.arg(sort_order)::text = 'desc' THEN name END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'order' AND sqlc.arg(sort_order)::text = 'asc' THEN "order" END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'order' AND sqlc.arg(sort_order)::text = 'desc' THEN "order" END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'asc' THEN id END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'desc' THEN id END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'created_at' AND sqlc.arg(sort_order)::text = 'asc' THEN created_at END ASC,
  created_at DESC,
  id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountSizes :one
SELECT COUNT(*) FROM sizes
WHERE (
  sqlc.arg(search)::text = ''
  OR name ILIKE '%' || sqlc.arg(search)::text || '%'
  OR type ILIKE '%' || sqlc.arg(search)::text || '%'
  OR "order" ILIKE '%' || sqlc.arg(search)::text || '%'
);

-- name: UpdateSize :one
UPDATE sizes
SET name = $2, type = $3, "order" = $4, version = version + 1
WHERE id = $1
RETURNING *;

-- name: DeleteSize :exec
DELETE FROM sizes
WHERE id = $1;
