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
ORDER BY id DESC
LIMIT $1
OFFSET $2;

-- name: CountExpenses :one
SELECT COUNT(*) FROM expenses;

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
