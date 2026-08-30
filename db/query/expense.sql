-- name: CreateExpense :one
INSERT INTO expenses (
  description,
  category_id,
  amount,
  currency_code,
  recurring_expense_id
)
VALUES ( $1, $2, $3, $4, $5 )
RETURNING *;

-- name: GetExpense :one
SELECT * FROM expenses
WHERE id = $1 LIMIT 1;

-- name: ListExpenses :many
SELECT * FROM expenses
WHERE (sqlc.narg(category_id)::bigint IS NULL OR category_id = sqlc.narg(category_id))
  AND (sqlc.narg(search)::text IS NULL OR description ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY id DESC
LIMIT $1
OFFSET $2;

-- name: CountExpenses :one
SELECT COUNT(*) FROM expenses
WHERE (sqlc.narg(category_id)::bigint IS NULL OR category_id = sqlc.narg(category_id))
  AND (sqlc.narg(search)::text IS NULL OR description ILIKE '%' || sqlc.narg(search) || '%');

-- name: UpdateExpense :one
UPDATE expenses
SET description = $2,
    category_id = $3,
    amount = $4,
    currency_code = $5
WHERE id = $1
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses
WHERE id = $1;

-- name: DeleteExpenses :execrows
DELETE FROM expenses
WHERE id = ANY(sqlc.arg(ids)::bigint[]);
