-- name: CreateBranchShift :one
INSERT INTO branch_shifts (
  branch_id,
  client_ref,
  branch_shift_id,
  opened_at,
  closed_at,
  closing_person_name,
  opening_float_usd,
  opening_float_lbp,
  expected_usd,
  expected_lbp,
  expected_visa,
  expected_whish,
  counted_usd,
  counted_lbp,
  counted_visa,
  counted_whish,
  variance_usd,
  variance_lbp,
  variance_visa,
  variance_whish,
  occurred_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
)
RETURNING *;

-- name: GetBranchShiftByClientRef :one
SELECT * FROM branch_shifts
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: ListBranchShifts :many
-- sqlc.narg(branch_id) is nullable: NULL means "all branches", matching
-- ListBranchExpenses' admin-filter convention.
SELECT * FROM branch_shifts
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchShifts :one
SELECT COUNT(*) FROM branch_shifts
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id);
