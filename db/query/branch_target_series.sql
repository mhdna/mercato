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

-- name: GetBranchTargetSeries :one
SELECT * FROM branch_target_series
WHERE id = $1 LIMIT 1;

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

-- name: UpdateBranchTargetSeries :one
-- Only affects future periods -- see CreateGeneratedBranchTarget's note on
-- generation snapshotting a series' fields, never joining them live.
UPDATE branch_target_series
SET target_amount = $2,
    color = $3,
    start_day = $4,
    interval_count = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteBranchTargetSeries :one
DELETE FROM branch_target_series
WHERE id = $1
RETURNING *;
