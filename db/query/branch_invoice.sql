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
  occurred_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

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

-- name: ListDailyIncome :many
-- Sums grand_total across both kinds ('sales' and 'return') rather than
-- treating them separately: a branch's return/exchange grand_total is
-- already the net signed cash effect it reported (negative for a refund,
-- positive for an even-or-up exchange -- see tx_exchange.go's netDifference
-- in kashi-pos), so a plain sum already nets returns against sales
-- correctly without this query needing to know or guess that sign
-- convention itself.
SELECT
  (occurred_at AT TIME ZONE 'UTC')::date AS day,
  SUM(grand_total)::bigint AS total
FROM branch_invoices
WHERE EXTRACT(YEAR FROM occurred_at) = sqlc.arg(year)::int
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
GROUP BY day
ORDER BY day;
