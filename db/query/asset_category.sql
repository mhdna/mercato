-- name: CreateAssetCategory :one
INSERT INTO asset_categories (
  name,
  is_active,
  icon,
  color
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetAssetCategory :one
SELECT * FROM asset_categories
WHERE id = $1 LIMIT 1;

-- name: ListAssetCategories :many
SELECT * FROM asset_categories
ORDER BY id;

-- name: UpdateAssetCategory :one
UPDATE asset_categories
SET name = $2,
    is_active = $3,
    icon = $4,
    color = $5
WHERE id = $1
RETURNING *;

-- name: DeleteAssetCategory :exec
DELETE FROM asset_categories
WHERE id = $1;
