-- name: CreateBranch :one
INSERT INTO branches (
  name,
  code,
  api_key_hash
) VALUES ( $1, $2, $3 )
RETURNING *;

-- name: GetBranchByCode :one
SELECT * FROM branches
WHERE code = $1 LIMIT 1;

-- name: UpdateBranchLastSeenAt :exec
UPDATE branches
SET last_seen_at = now()
WHERE id = $1;

-- name: ListBranches :many
SELECT * FROM branches
ORDER BY name;

-- name: GetBranch :one
SELECT * FROM branches
WHERE id = $1 LIMIT 1;

-- name: SetBranchActive :exec
UPDATE branches
SET is_active = $2
WHERE id = $1;

-- name: UpdateBranchAPIKeyHash :exec
UPDATE branches
SET api_key_hash = $2
WHERE id = $1;
