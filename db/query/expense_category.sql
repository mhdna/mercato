-- name: CreateExpenseCategory :one
INSERT INTO expense_categories (
  name,
  is_active
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetExpenseCategory :one
SELECT * FROM expense_categories
WHERE id = $1 LIMIT 1;

-- name: ListExpenseCategories :many
SELECT * FROM expense_categories
ORDER BY id;

-- name: UpdateExpenseCategory :one
UPDATE expense_categories
SET name = $2,
    is_active = $3
WHERE id = $1
RETURNING *;

-- name: DeleteExpenseCategory :exec
DELETE FROM expense_categories
WHERE id = $1;
