-- name: GetBranchVisitorEventByClientRef :one
SELECT * FROM branch_visitor_events
WHERE branch_id = $1 AND client_ref = $2
LIMIT 1;

-- name: CreateBranchVisitorEvent :one
INSERT INTO branch_visitor_events (
  branch_id,
  client_ref,
  day,
  direction,
  occurred_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListBranchVisitorDays :many
-- One row per (branch, local day): gross customers-in vs customers-out,
-- newest day first. sqlc.narg(branch_id) NULL means "all branches",
-- matching ListBranchShifts' admin-filter convention.
SELECT
  branch_id,
  day,
  COUNT(*) FILTER (WHERE direction = 'in')  AS in_count,
  COUNT(*) FILTER (WHERE direction = 'out') AS out_count
FROM branch_visitor_events
WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
GROUP BY branch_id, day
ORDER BY day DESC, branch_id
LIMIT $1 OFFSET $2;

-- name: CountBranchVisitorDays :one
SELECT COUNT(*) FROM (
  SELECT 1
  FROM branch_visitor_events
  WHERE sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id)
  GROUP BY branch_id, day
) t;

-- name: ListBranchVisitorDaysRange :many
-- Every (branch, day) rollup within an inclusive "YYYY-MM-DD" day range,
-- oldest first. No pagination: footfall data is one row per branch per day,
-- so even a year across every branch is a small result the Footfall page
-- slices into months / branch comparisons / weekday profiles client-side.
SELECT
  branch_id,
  day,
  COUNT(*) FILTER (WHERE direction = 'in')  AS in_count,
  COUNT(*) FILTER (WHERE direction = 'out') AS out_count
FROM branch_visitor_events
WHERE day >= sqlc.arg(from_day)
  AND day <= sqlc.arg(to_day)
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
GROUP BY branch_id, day
ORDER BY day, branch_id;

-- name: CountBranchVisitorDayByDirection :one
-- Running gross tally for one branch, one local day, one direction -- used
-- to tell the admin toast "customer in -- 42 in today".
SELECT COUNT(*)
FROM branch_visitor_events
WHERE branch_id = $1 AND day = $2 AND direction = $3;
