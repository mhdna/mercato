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