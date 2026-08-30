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

-- name: UpdateLoan :one
-- Central loans only -- branch-origin loans are a synced record of what a
-- branch reported and stay read-only here (WHERE origin filter makes a
-- branch-loan id return no rows, surfaced as a 404 by the handler).
UPDATE loans
SET description = $2,
    category_id = $3,
    amount = $4,
    currency_code = $5
WHERE id = $1 AND origin = 'central_loan'
RETURNING *;

-- name: DeleteLoan :exec
DELETE FROM loans
WHERE id = $1 AND origin = 'central_loan';

-- name: DeleteLoans :execrows
-- Central loans only, same read-only guard as DeleteLoan.
DELETE FROM loans
WHERE id = ANY(sqlc.arg(ids)::bigint[]) AND origin = 'central_loan';

-- name: ListLoans :many
-- sqlc.narg(branch_id)/sqlc.narg(origin) are nullable: NULL means "no
-- filter", matching ListBranchExpenses' admin-filter convention.
-- paid_amount is the running total of loan_payments against the loan;
-- status is derived from it (never stored) -- 'paid' once payments cover
-- the loan amount, 'partial' while some but not all is covered, else
-- 'unpaid'. sqlc.narg(status) filters on that same derived value.
SELECT
  loans.*,
  COALESCE(pay.paid_amount, 0)::bigint AS paid_amount,
  (CASE
    WHEN COALESCE(pay.paid_amount, 0) >= loans.amount THEN 'paid'
    WHEN COALESCE(pay.paid_amount, 0) > 0 THEN 'partial'
    ELSE 'unpaid'
  END)::text AS status
FROM loans
LEFT JOIN (
  SELECT loan_id, SUM(amount) AS paid_amount
  FROM loan_payments
  GROUP BY loan_id
) pay ON pay.loan_id = loans.id
WHERE (sqlc.narg(branch_id)::bigint IS NULL OR loans.branch_id = sqlc.narg(branch_id))
  AND (sqlc.narg(origin)::text IS NULL OR loans.origin = sqlc.narg(origin))
  AND (sqlc.narg(category_id)::bigint IS NULL OR loans.category_id = sqlc.narg(category_id))
  AND (sqlc.narg(search)::text IS NULL OR loans.description ILIKE '%' || sqlc.narg(search) || '%')
  AND (sqlc.narg(status)::text IS NULL OR sqlc.narg(status) = (CASE
    WHEN COALESCE(pay.paid_amount, 0) >= loans.amount THEN 'paid'
    WHEN COALESCE(pay.paid_amount, 0) > 0 THEN 'partial'
    ELSE 'unpaid'
  END))
ORDER BY loans.id DESC
LIMIT $1 OFFSET $2;

-- name: CountLoans :one
SELECT COUNT(*)
FROM loans
LEFT JOIN (
  SELECT loan_id, SUM(amount) AS paid_amount
  FROM loan_payments
  GROUP BY loan_id
) pay ON pay.loan_id = loans.id
WHERE (sqlc.narg(branch_id)::bigint IS NULL OR loans.branch_id = sqlc.narg(branch_id))
  AND (sqlc.narg(origin)::text IS NULL OR loans.origin = sqlc.narg(origin))
  AND (sqlc.narg(category_id)::bigint IS NULL OR loans.category_id = sqlc.narg(category_id))
  AND (sqlc.narg(search)::text IS NULL OR loans.description ILIKE '%' || sqlc.narg(search) || '%')
  AND (sqlc.narg(status)::text IS NULL OR sqlc.narg(status) = (CASE
    WHEN COALESCE(pay.paid_amount, 0) >= loans.amount THEN 'paid'
    WHEN COALESCE(pay.paid_amount, 0) > 0 THEN 'partial'
    ELSE 'unpaid'
  END));
