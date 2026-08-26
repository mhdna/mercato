-- name: CreateBranchExpenseUploadToken :one
INSERT INTO branch_expense_upload_tokens (
  branch_expense_id,
  token,
  expires_at
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetBranchExpenseUploadTokenByToken :one
SELECT * FROM branch_expense_upload_tokens
WHERE token = $1
LIMIT 1;

-- name: MarkBranchExpenseUploadTokenUsed :exec
UPDATE branch_expense_upload_tokens
SET used_at = now()
WHERE id = $1;

-- name: CreateBranchExpenseImage :one
INSERT INTO branch_expense_images (
  branch_expense_id,
  file_path,
  content_type,
  size_bytes
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetBranchExpenseImage :one
SELECT * FROM branch_expense_images
WHERE id = $1
LIMIT 1;

-- name: ListBranchExpenseImagesForExpense :many
SELECT * FROM branch_expense_images
WHERE branch_expense_id = $1
ORDER BY id DESC;

-- Storage-explorer feed: every uploaded image, newest first, joined with
-- its expense and branch so the admin UI never needs a second round trip
-- per row.
-- name: ListBranchExpenseImages :many
SELECT
  branch_expense_images.*,
  branch_expenses.description AS expense_description,
  branch_expenses.branch_id AS expense_branch_id,
  branches.name AS branch_name
FROM branch_expense_images
JOIN branch_expenses ON branch_expenses.id = branch_expense_images.branch_expense_id
JOIN branches ON branches.id = branch_expenses.branch_id
ORDER BY branch_expense_images.id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchExpenseImages :one
SELECT COUNT(*) FROM branch_expense_images;
