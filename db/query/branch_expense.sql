-- name: CreateBranchExpense :one
INSERT INTO branch_expenses (
  branch_id,
  client_ref,
  description,
  category,
  amount,
  currency_code,
  branch_cashbox_account_id,
  branch_shift_id,
  occurred_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetBranchExpenseByClientRef :one
SELECT * FROM branch_expenses
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: ListBranchExpenses :many
-- sqlc.narg(branch_id) is nullable: NULL means "all branches", matching
-- ListBranchInvoices' admin-filter convention.
SELECT * FROM branch_expenses
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchExpenses :one
SELECT COUNT(*) FROM branch_expenses
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id);
