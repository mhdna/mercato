-- name: CreateExpenseCategory :one
INSERT INTO expense_categories (
  name,
  is_active,
  icon,
  color,
  scope
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetExpenseCategory :one
SELECT * FROM expense_categories
WHERE id = $1 LIMIT 1;

-- name: ListExpenseCategories :many
SELECT * FROM expense_categories
WHERE (sqlc.narg(scope)::text IS NULL OR scope = sqlc.narg(scope))
ORDER BY id;

-- name: ListExpenseCategoriesUpdatedSince :many
-- Only branch-scoped categories are synced down to kashi-pos.
SELECT * FROM expense_categories
WHERE scope = 'branch' AND updated_at > $1
ORDER BY updated_at;

-- name: UpdateExpenseCategory :one
UPDATE expense_categories
SET name = $2,
    is_active = $3,
    icon = $4,
    color = $5,
    scope = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteExpenseCategory :exec
DELETE FROM expense_categories
WHERE id = $1;
