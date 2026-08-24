-- name: CreateClient :one
INSERT INTO clients (
  name,
  phone
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetClient :one
SELECT * FROM clients
WHERE id = $1 LIMIT 1;

-- name: ListClients :many
SELECT * FROM clients
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: UpdateClient :one
UPDATE clients 
  SET name = $2,
  phone = $3
WHERE id = $1
RETURNING *;

-- name: DeleteClient :exec
DELETE FROM clients
WHERE id = $1;

-- name: AddClientLoyaltyPoints :exec
UPDATE clients 
  SET total_loyalty_points = total_loyalty_points + $2,
  valid_loyalty_points = valid_loyalty_points + $3
WHERE id = $1;
-- name: CountClients :one
SELECT COUNT(*) FROM clients;
