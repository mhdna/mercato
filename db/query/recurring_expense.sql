-- name: CreateRecurringExpense :one
INSERT INTO recurring_expenses (
  description,
  category,
  amount,
  currency_code,
  interval_unit,
  interval_count,
  next_due_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetRecurringExpense :one
SELECT * FROM recurring_expenses WHERE id = $1;

-- name: ListRecurringExpenses :many
SELECT * FROM recurring_expenses ORDER BY id DESC;

-- name: SetRecurringExpenseActive :one
UPDATE recurring_expenses SET active = $2 WHERE id = $1 RETURNING *;

-- name: ListDueRecurringExpenses :many
SELECT * FROM recurring_expenses
WHERE active AND next_due_at <= $1
ORDER BY id;

-- name: UpdateRecurringExpenseNextDue :one
UPDATE recurring_expenses SET next_due_at = $2 WHERE id = $1 RETURNING *;
