-- name: CreateAsset :one
INSERT INTO assets (
  name,
  code,
  category_id,
  bought_at
) VALUES ( $1, $2, $3, $4 )
RETURNING *;

-- name: GetAsset :one
SELECT * FROM assets
WHERE id = $1 LIMIT 1;

-- name: ListAssetsPage :many
SELECT * FROM assets
WHERE (sqlc.narg(category_id)::bigint IS NULL OR category_id = sqlc.narg(category_id))
  AND (
    sqlc.arg(search)::text = ''
    OR name ILIKE '%' || sqlc.arg(search)::text || '%'
    OR code ILIKE '%' || sqlc.arg(search)::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'name' AND sqlc.arg(sort_order)::text = 'asc' THEN name END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'name' AND sqlc.arg(sort_order)::text = 'desc' THEN name END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'code' AND sqlc.arg(sort_order)::text = 'asc' THEN code END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'code' AND sqlc.arg(sort_order)::text = 'desc' THEN code END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'bought_at' AND sqlc.arg(sort_order)::text = 'asc' THEN bought_at END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'bought_at' AND sqlc.arg(sort_order)::text = 'desc' THEN bought_at END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'asc' THEN id END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'desc' THEN id END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'created_at' AND sqlc.arg(sort_order)::text = 'asc' THEN created_at END ASC,
  created_at DESC,
  id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountAssets :one
SELECT COUNT(*) FROM assets
WHERE (sqlc.narg(category_id)::bigint IS NULL OR category_id = sqlc.narg(category_id))
  AND (
    sqlc.arg(search)::text = ''
    OR name ILIKE '%' || sqlc.arg(search)::text || '%'
    OR code ILIKE '%' || sqlc.arg(search)::text || '%'
  );

-- name: UpdateAsset :one
UPDATE assets
  SET name = $2,
  code = $3,
  category_id = $4,
  bought_at = $5,
  version = version + 1
WHERE id = $1
RETURNING *;

-- name: DeleteAsset :exec
DELETE FROM assets
WHERE id = $1;

-- name: DeleteAssets :execrows
DELETE FROM assets
WHERE id = ANY(sqlc.arg(ids)::bigint[]);
