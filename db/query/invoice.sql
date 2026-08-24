-- name: CreateInvoice :one
INSERT INTO invoices (
  cashbox_id,
  shift_id,
  invoice_code,
  invoice_index,
  year,
  client_id,
  inventory_id,
  discount,
  subtotal,
  discounted_total,
  grand_total,
  invoice_type_id
)
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: AddInvoiceProduct :one
INSERT INTO invoice_products (
  invoice_id,
  product_id,
  unit_price,
  line_total,
  discount,
  quantity
)
VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: GetInvoice :one
SELECT * FROM invoices
WHERE id = $1 LIMIT 1;

-- name: CreateSalesInvoice :one
INSERT INTO sales_invoices (invoice_id)
VALUES ($1)
RETURNING *;

-- name: GetSalesInvoice :one
SELECT * FROM sales_invoices
WHERE invoice_id = $1 LIMIT 1;

-- name: CreateReturnInvoice :one
INSERT INTO return_invoices (invoice_id, sales_invoice_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetReturnInvoice :one
SELECT * FROM return_invoices
WHERE invoice_id = $1 LIMIT 1;

-- name: ListInvoices :many
SELECT *
FROM invoices
ORDER BY created_at
DESC
LIMIT $1
OFFSET $2;

-- name: IncrementInvoicesIndex :one
INSERT INTO invoice_indexes  (year, cashbox_id, type, invoice_type_id, last_index)
VALUES ($1, $2, $3, $4, 1)
ON CONFLICT (year, cashbox_id, invoice_type_id)
DO UPDATE SET last_index = invoice_indexes.last_index + 1, type = $3
RETURNING last_index;

-- name: DecrementInvoicesIndex :one
INSERT INTO invoice_indexes  (year, cashbox_id, type, invoice_type_id, last_index)
VALUES ($1, $2, $3, $4, 1)
ON CONFLICT (year, cashbox_id, invoice_type_id)
DO UPDATE SET last_index = invoice_indexes.last_index - 1, type = $3
RETURNING last_index;

-- name: CountInvoices :one
SELECT COUNT(*) FROM invoices;
