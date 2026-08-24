-- name: CreateBranchCommand :one
INSERT INTO branch_commands (
  branch_id, type, payload, issued_by
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetBranchCommand :one
SELECT * FROM branch_commands WHERE id = $1;

-- name: ListPendingBranchCommands :many
SELECT * FROM branch_commands
WHERE branch_id = $1 AND status = 'pending'
ORDER BY id;

-- name: CompleteBranchCommand :one
UPDATE branch_commands
SET status = $2, result = $3, error = $4, completed_at = now()
WHERE id = $1
RETURNING *;

-- name: ListBranchCommands :many
SELECT * FROM branch_commands
WHERE branch_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;
