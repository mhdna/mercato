-- name: CreateInvoiceType :one
INSERT INTO invoice_types (
  name,
  code,
  is_default,
  is_active
)
VALUES ( $1, $2, $3, $4 )
RETURNING *;

-- name: GetInvoiceType :one
SELECT * FROM invoice_types
WHERE id = $1 LIMIT 1;

-- name: GetDefaultInvoiceType :one
SELECT * FROM invoice_types
WHERE is_default = true LIMIT 1;

-- name: ListInvoiceTypes :many
SELECT * FROM invoice_types
ORDER BY id;

-- name: UpdateInvoiceType :one
UPDATE invoice_types
SET name = $2,
    code = $3,
    is_active = $4
WHERE id = $1
RETURNING *;
