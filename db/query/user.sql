-- name: CreateUser :one
INSERT INTO users (
  name,
  email,
  password_hash,
  activated
) VALUES ( $1, $2, $3, $4 )
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUser :exec
UPDATE users
  SET name = $2,
  email = $3,
  password_hash = $4,
  activated = $5
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
