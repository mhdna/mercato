-- name: CreateBranchUser :one
INSERT INTO branch_users (branch_id, username, role, akuvox_user_id, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetBranchUser :one
SELECT * FROM branch_users WHERE id = $1;

-- name: ListBranchUsersForBranch :many
SELECT * FROM branch_users
WHERE branch_id = $1
ORDER BY username;

-- name: UpdateBranchUser :one
UPDATE branch_users
SET username = $2,
    role = $3,
    akuvox_user_id = $4,
    is_active = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetBranchUserActive :one
UPDATE branch_users
SET is_active = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteBranchUser :exec
DELETE FROM branch_users WHERE id = $1;

-- Branch-scoped catch-up feed, same contract as ListSalespersonsUpdatedSince.
-- name: ListBranchUsersUpdatedSince :many
SELECT * FROM branch_users
WHERE branch_id = $1 AND updated_at > $2
ORDER BY updated_at;
