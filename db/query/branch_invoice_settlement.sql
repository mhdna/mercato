-- name: CreateBranchInvoiceSettlement :one
INSERT INTO branch_invoice_settlements (
  branch_id,
  client_ref,
  sale_client_ref,
  grand_total,
  occurred_at,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: BranchShiftIsClosed :one
-- A branch_shifts row exists only once kashi-pos has reported that shift's
-- close, so its mere presence is the "this shift is closed" signal --
-- closed_at itself can be NULL for an older till build that omitted it.
SELECT EXISTS (
  SELECT 1 FROM branch_shifts
  WHERE branch_id = $1 AND branch_shift_id = $2
) AS closed;

-- name: GetBranchInvoiceSettlementByClientRef :one
SELECT * FROM branch_invoice_settlements
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: CreateBranchInvoiceSettlementPayment :one
INSERT INTO branch_invoice_settlement_payments (
  branch_invoice_settlement_id,
  account_name,
  amount
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: ListBranchInvoiceSettlementPayments :many
SELECT * FROM branch_invoice_settlement_payments
WHERE branch_invoice_settlement_id = $1
ORDER BY id;

-- name: ListBranchInvoiceSettlements :many
SELECT * FROM branch_invoice_settlements
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchInvoiceSettlements :one
SELECT COUNT(*) FROM branch_invoice_settlements
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id);

-- name: DeleteBranchInvoicePaymentsForInvoice :exec
DELETE FROM branch_invoice_payments
WHERE branch_invoice_id = $1;
