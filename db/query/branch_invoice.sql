-- name: CreateBranchInvoice :one
INSERT INTO branch_invoices (
  branch_id,
  client_ref,
  kind,
  branch_invoice_code,
  branch_cashbox_account_id,
  branch_shift_id,
  branch_inventory_id,
  branch_client_id,
  related_client_ref,
  discount,
  subtotal,
  discounted_total,
  grand_total,
  loyalty_points_delta,
  occurred_at,
  salesperson_name
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: CreateBranchInvoicePayment :one
INSERT INTO branch_invoice_payments (branch_invoice_id, account_name, amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListBranchInvoicePayments :many
SELECT * FROM branch_invoice_payments
WHERE branch_invoice_id = $1
ORDER BY id;

-- name: GetBranchInvoiceByClientRef :one
SELECT * FROM branch_invoices
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: CreateBranchInvoiceItem :one
INSERT INTO branch_invoice_items (
  branch_invoice_id,
  branch_product_id,
  unit_price,
  line_total,
  discount,
  quantity
) VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: ListBranchInvoiceItems :many
SELECT * FROM branch_invoice_items
WHERE branch_invoice_id = $1;

-- name: GetBranchInvoice :one
SELECT * FROM branch_invoices WHERE id = $1;

-- name: ListBranchInvoices :many
-- sqlc.narg(branch_id) is nullable: NULL means "all branches", matching how
-- the admin UI's branch filter works (a dropdown with an "All branches"
-- option, not a required selection).
SELECT * FROM branch_invoices
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchInvoices :one
SELECT COUNT(*) FROM branch_invoices
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id);

-- name: ListBranchInvoicesPage :many
SELECT * FROM branch_invoices
WHERE (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
  AND (
    sqlc.arg(search)::text = ''
    OR branch_invoice_code ILIKE '%' || sqlc.arg(search)::text || '%'
    OR salesperson_name ILIKE '%' || sqlc.arg(search)::text || '%'
  )
ORDER BY id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountBranchInvoicesFiltered :one
SELECT COUNT(*) FROM branch_invoices
WHERE (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
  AND (
    sqlc.arg(search)::text = ''
    OR branch_invoice_code ILIKE '%' || sqlc.arg(search)::text || '%'
    OR salesperson_name ILIKE '%' || sqlc.arg(search)::text || '%'
  );

-- name: ListDailyIncome :many
-- Sums grand_total across both kinds ('sales' and 'return') rather than
-- treating them separately: a branch's return/exchange grand_total is
-- already the net signed cash effect it reported (negative for a refund,
-- positive for an even-or-up exchange -- see tx_exchange.go's netDifference
-- in kashi-pos), so a plain sum already nets returns against sales
-- correctly without this query needing to know or guess that sign
-- convention itself.
-- The year filter is a >=/< range against a fixed pair of dates rather than
-- EXTRACT(YEAR FROM occurred_at) = $1 -- EXTRACT on every row can't use a
-- plain btree index on occurred_at (idx_branch_invoices_occurred_at), while
-- a range comparison can.
SELECT
  (occurred_at AT TIME ZONE 'UTC')::date AS day,
  SUM(grand_total)::bigint AS total
FROM branch_invoices
WHERE occurred_at >= make_date(sqlc.arg(year)::int, 1, 1)
  AND occurred_at < make_date(sqlc.arg(year)::int + 1, 1, 1)
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
GROUP BY day
ORDER BY day;
