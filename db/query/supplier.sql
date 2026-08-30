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
WHERE (
  BTRIM(sqlc.arg(search)::text) = ''
  OR name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR phone ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR country ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR address ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
)
ORDER BY id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountSuppliersFiltered :one
SELECT COUNT(*) FROM suppliers
WHERE (
  BTRIM(sqlc.arg(search)::text) = ''
  OR name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR phone ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR country ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  OR address ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
);
