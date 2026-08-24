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
