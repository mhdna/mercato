-- name: CreateLoanPayment :one
INSERT INTO loan_payments (
  loan_id,
  amount,
  currency_code,
  note
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetLoanPayment :one
SELECT * FROM loan_payments
WHERE id = $1 LIMIT 1;

-- name: ListLoanPayments :many
-- sqlc.narg(loan_id) is nullable: NULL means "all loans".
SELECT * FROM loan_payments
WHERE sqlc.narg(loan_id)::bigint IS NULL OR loan_id = sqlc.narg(loan_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountLoanPayments :one
SELECT COUNT(*) FROM loan_payments
WHERE sqlc.narg(loan_id)::bigint IS NULL OR loan_id = sqlc.narg(loan_id);

-- name: DeleteLoanPayment :exec
DELETE FROM loan_payments
WHERE id = $1;
