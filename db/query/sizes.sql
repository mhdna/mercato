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
