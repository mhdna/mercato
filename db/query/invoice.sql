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
  invoice_type_id,
  salesperson_id,
  loyalty_points_delta
)
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: AddInvoiceProduct :one
INSERT INTO invoice_products (
  invoice_id,
  product_id,
  variant_id,
  unit_price,
  line_total,
  discount,
  quantity,
  unit_cost
)
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8 )
RETURNING *;

-- name: GetInvoice :one
SELECT * FROM invoices
WHERE id = $1 LIMIT 1;

-- name: ListInvoiceProductsByInvoice :many
SELECT
  invoice_products.*,
  products.name AS product_name,
  products.code AS product_code,
  COALESCE(pv.barcode, '') AS variant_barcode,
  COALESCE(c.name, '') AS color_name,
  COALESCE(s.name, '') AS size_name
FROM invoice_products
JOIN products ON products.id = invoice_products.product_id
LEFT JOIN product_variants pv ON pv.id = invoice_products.variant_id
LEFT JOIN colors c ON c.id = pv.color_id
LEFT JOIN sizes  s ON s.id = pv.size_id
WHERE invoice_products.invoice_id = $1
ORDER BY invoice_products.product_id;

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

-- name: ListInvoicesPage :many
SELECT *
FROM invoices
WHERE (
  sqlc.arg(search)::text = ''
  OR invoice_code ILIKE '%' || sqlc.arg(search)::text || '%'
)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountInvoicesFiltered :one
SELECT COUNT(*)
FROM invoices
WHERE (
  sqlc.arg(search)::text = ''
  OR invoice_code ILIKE '%' || sqlc.arg(search)::text || '%'
);

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
