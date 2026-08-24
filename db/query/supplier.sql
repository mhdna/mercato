-- name: CreateSupplier :one
INSERT INTO suppliers (
  name,
  phone,
  country,
  address,
  address_latitude,
  address_longitude
) 
VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: GetSupplier :one
SELECT * FROM suppliers
WHERE id = $1 LIMIT 1;

-- TOOD: add UpdateSupplier

-- name: ListSuppliers :many
SELECT * FROM suppliers
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CountSuppliers :one
SELECT COUNT(*) FROM suppliers;