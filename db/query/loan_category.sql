-- name: CreateLoanCategory :one
INSERT INTO loan_categories (
  name,
  is_active,
  icon,
  color,
  scope
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetLoanCategory :one
SELECT * FROM loan_categories
WHERE id = $1 LIMIT 1;

-- name: ListLoanCategories :many
SELECT * FROM loan_categories
WHERE (sqlc.narg(scope)::text IS NULL OR scope = sqlc.narg(scope))
ORDER BY id;

-- name: ListLoanCategoriesUpdatedSince :many
-- Only branch-scoped categories are synced down to kashi-pos -- central
-- ones are managed in kashi only.
SELECT * FROM loan_categories
WHERE scope = 'branch' AND updated_at > $1
ORDER BY updated_at;

-- name: UpdateLoanCategory :one
UPDATE loan_categories
SET name = $2,
    is_active = $3,
    icon = $4,
    color = $5,
    scope = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteLoanCategory :exec
DELETE FROM loan_categories
WHERE id = $1;
