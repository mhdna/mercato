-- name: CreateBranchTargetSeries :one
INSERT INTO branch_target_series (
  branch_id,
  target_amount,
  color,
  start_day,
  interval_count
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListBranchTargetSeriesForBranch :many
SELECT * FROM branch_target_series
WHERE branch_id = $1
ORDER BY id DESC;

-- name: ListActiveBranchTargetSeries :many
SELECT * FROM branch_target_series
WHERE active
ORDER BY id;

-- name: SetBranchTargetSeriesActive :one
UPDATE branch_target_series
SET active = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;
