-- name: CreateCentralLoan :one
INSERT INTO loans (
  origin,
  description,
  category_id,
  amount,
  currency_code
) VALUES (
  'central_loan', $1, $2, $3, $4
)
RETURNING *;

-- name: CreateBranchLoan :one
INSERT INTO loans (
  origin,
  branch_id,
  client_ref,
  description,
  category_id,
  amount,
  currency_code,
  branch_cashbox_account_id,
  branch_shift_id,
  occurred_at
) VALUES (
  'branch_loan', $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetLoan :one
SELECT * FROM loans
WHERE id = $1 LIMIT 1;

-- name: GetBranchLoanByClientRef :one
SELECT * FROM loans
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: ListLoans :many
-- sqlc.narg(branch_id)/sqlc.narg(origin) are nullable: NULL means "no
-- filter", matching ListBranchExpenses' admin-filter convention.
SELECT * FROM loans
WHERE (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
  AND (sqlc.narg(origin)::text IS NULL OR origin = sqlc.narg(origin))
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountLoans :one
SELECT COUNT(*) FROM loans
WHERE (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
  AND (sqlc.narg(origin)::text IS NULL OR origin = sqlc.narg(origin));
