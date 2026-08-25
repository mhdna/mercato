-- name: CreateBranchTarget :one
INSERT INTO branch_targets (
  branch_id,
  date_from,
  date_to,
  target_amount,
  color
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetBranchTarget :one
SELECT * FROM branch_targets
WHERE id = $1 LIMIT 1;

-- name: ListBranchTargetsForBranch :many
SELECT * FROM branch_targets
WHERE branch_id = $1
ORDER BY target_amount;

-- name: UpdateBranchTarget :one
UPDATE branch_targets
SET date_from = $2,
    date_to = $3,
    target_amount = $4,
    color = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteBranchTarget :exec
DELETE FROM branch_targets
WHERE id = $1;

-- name: ListBranchTargetsUpdatedSince :many
SELECT * FROM branch_targets
WHERE branch_id = $1 AND updated_at > $2
ORDER BY updated_at;

-- name: CreateGeneratedBranchTarget :one
-- Fires one period for a recurring series -- target_amount/color are
-- passed in (a snapshot of the series at generation time) rather than
-- joined live, so a later edit to the series template never rewrites a
-- period that's already in progress or past, same as recurring_expenses'
-- FireRecurringExpenseTx snapshotting its template's amount into the
-- created expense row.
INSERT INTO branch_targets (
  branch_id,
  date_from,
  date_to,
  target_amount,
  color,
  series_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetBranchTargetBySeriesAndStart :one
SELECT * FROM branch_targets
WHERE series_id = $1 AND date_from = $2
LIMIT 1;

-- name: SumBranchRevenueForRange :one
-- Net sales revenue: same grand_total-sum convention as ListDailyIncome in
-- branch_invoice.sql (returns already net out via their signed grand_total).
SELECT COALESCE(SUM(grand_total), 0)::bigint AS total
FROM branch_invoices
WHERE branch_id = sqlc.arg(branch_id)
  AND (occurred_at AT TIME ZONE 'UTC')::date BETWEEN sqlc.arg(date_from) AND sqlc.arg(date_to);
