-- name: CreateBranchAttendanceEvent :one
INSERT INTO branch_attendance_events (
  branch_id,
  client_ref,
  salesperson_name,
  attendance_user_id,
  event_date,
  event_time,
  event_at,
  type,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetBranchAttendanceEventByClientRef :one
SELECT * FROM branch_attendance_events
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: GetBranchAttendanceEvent :one
SELECT * FROM branch_attendance_events
WHERE id = $1 LIMIT 1;

-- name: UpdateBranchAttendanceEvent :one
UPDATE branch_attendance_events
SET salesperson_name = $2,
    attendance_user_id = $3,
    event_date = $4,
    event_time = $5,
    event_at = $6,
    type = $7,
    status = $8
WHERE id = $1
RETURNING *;

-- name: DeleteBranchAttendanceEvent :exec
DELETE FROM branch_attendance_events WHERE id = $1;

-- name: ListBranchAttendanceEvents :many
SELECT * FROM branch_attendance_events
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchAttendanceEvents :one
SELECT COUNT(*) FROM branch_attendance_events
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id);

-- name: CreateBranchAttendanceChange :one
INSERT INTO branch_attendance_changes (
  branch_id,
  client_ref,
  kind,
  salesperson_name,
  attendance_date,
  original_time,
  requested_time,
  requested_type,
  action,
  status,
  note,
  actor,
  occurred_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetBranchAttendanceChangeByClientRef :one
SELECT * FROM branch_attendance_changes
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: ListBranchAttendanceChanges :many
SELECT * FROM branch_attendance_changes
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountBranchAttendanceChanges :one
SELECT COUNT(*) FROM branch_attendance_changes
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id);
