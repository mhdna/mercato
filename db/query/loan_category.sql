-- name: CreateLoanCategory :one
INSERT INTO loan_categories (
  name,
  is_active
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetLoanCategory :one
SELECT * FROM loan_categories
WHERE id = $1 LIMIT 1;

-- name: ListLoanCategories :many
SELECT * FROM loan_categories
ORDER BY id;

-- name: UpdateLoanCategory :one
UPDATE loan_categories
SET name = $2,
    is_active = $3
WHERE id = $1
RETURNING *;

-- name: DeleteLoanCategory :exec
DELETE FROM loan_categories
WHERE id = $1;
