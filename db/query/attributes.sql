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
ON a.name = av.attribute
ORDER BY value
LIMIT $1
OFFSET $2;

-- name: UpdateAttributeValue :one
UPDATE attributes_values
SET value = $2
WHERE id = $1
RETURNING *;
