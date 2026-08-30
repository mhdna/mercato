-- name: GetAttribute :one
Select *
from attributes
WHERe name = $1;

-- name: ListAttributes :many
SELECT *
FROM attributes
ORDER BY name;

-- name: UpsertAttributeValue :one
INSERT INTO attributes_values (attribute_id, value)
VALUES ($1, $2)
ON CONFLICT (attribute_id, value) DO UPDATE SET value = $2
RETURNING *; -- upserting instead of inserting makes product creation with attributes easier

-- name: GetAttributeValue :one
SELECT * FROM attributes_values
WHERE id = $1;


-- name: ListAttributeValues :many
SELECT a.*, av.*
FROM attributes a
INNER JOIN attributes_values av
ON a.id = av.attribute_id
ORDER BY value
LIMIT $1
OFFSET $2;

-- name: ListAttributeValuesPage :many
SELECT av.id, av.attribute_id, av.value, av.created_at, a.name AS attribute_name
FROM attributes_values av
INNER JOIN attributes a ON a.id = av.attribute_id
WHERE a.name = sqlc.arg(attribute_name)::text
  AND (
    sqlc.arg(search)::text = ''
    OR av.value ILIKE '%' || sqlc.arg(search)::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'value' AND sqlc.arg(sort_order)::text = 'asc' THEN av.value END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'value' AND sqlc.arg(sort_order)::text = 'desc' THEN av.value END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'asc' THEN av.id END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'id' AND sqlc.arg(sort_order)::text = 'desc' THEN av.id END DESC,
  CASE WHEN sqlc.arg(sort_by)::text = 'created_at' AND sqlc.arg(sort_order)::text = 'asc' THEN av.created_at END ASC,
  av.created_at DESC,
  av.id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountAttributeValues :one
SELECT COUNT(*)
FROM attributes_values av
INNER JOIN attributes a ON a.id = av.attribute_id
WHERE a.name = sqlc.arg(attribute_name)::text
  AND (
    sqlc.arg(search)::text = ''
    OR av.value ILIKE '%' || sqlc.arg(search)::text || '%'
  );

-- name: ListAllAttributeValues :many
SELECT av.id, av.attribute_id, av.value, av.created_at, a.name AS attribute_name
FROM attributes_values av
INNER JOIN attributes a ON a.id = av.attribute_id
ORDER BY a.name, av.value;

-- name: UpdateAttributeValue :one
UPDATE attributes_values
SET value = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAttributeValue :exec
DELETE FROM attributes_values
WHERE id = $1;
