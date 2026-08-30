-- name: CreateColor :one
INSERT INTO colors (
  name,
  hex_value
) VALUES (
    $1, $2
) RETURNING *;

-- name: ListColors :many
SELECT * FROM colors
ORDER BY name;

-- name: ListColorsPage :many
SELECT * FROM colors
WHERE (
  sqlc.arg(search)::text = ''
  OR name ILIKE '%' || sqlc.arg(search)::text || '%'
  OR hex_value ILIKE '%' || sqlc.arg(search)::text || '%'
)
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'name' AND sqlc.arg(sort_order)::text = 'asc' THEN name END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'name' AND sqlc.arg(sort_order)::text = 'desc' THEN name END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'hex_value' AND sqlc.arg(sort_order)::text = 'asc' THEN hex_value END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'hex_value' AND sqlc.arg(sort_order)::text = 'desc' THEN hex_value END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'asc' THEN id END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'desc' THEN id END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'created_at' AND sqlc.arg(sort_order)::text = 'asc' THEN created_at END ASC,
  created_at DESC,
  id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountColors :one
SELECT COUNT(*) FROM colors
WHERE (
  sqlc.arg(search)::text = ''
  OR name ILIKE '%' || sqlc.arg(search)::text || '%'
  OR hex_value ILIKE '%' || sqlc.arg(search)::text || '%'
);

-- name: UpdateColor :one
UPDATE colors
SET name = $2, hex_value = $3, version = version + 1
WHERE id = $1
RETURNING *;

-- name: DeleteColor :exec
DELETE FROM colors
WHERE id = $1;
