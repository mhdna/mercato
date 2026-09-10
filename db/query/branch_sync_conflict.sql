-- name: UpsertBranchSyncConflict :one
-- Idempotent on (branch_id, entity, ref): a retried outbox push re-reports
-- the same conflict. A row an admin already resolved/dismissed is left
-- alone (nothing to re-open).
INSERT INTO branch_sync_conflicts (branch_id, entity, ref, kind, branch_payload, central_client_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (branch_id, entity, ref) DO UPDATE SET
    kind = EXCLUDED.kind,
    branch_payload = EXCLUDED.branch_payload,
    central_client_id = EXCLUDED.central_client_id,
    updated_at = now()
WHERE branch_sync_conflicts.status = 'open'
RETURNING *;

-- name: GetBranchSyncConflict :one
SELECT * FROM branch_sync_conflicts WHERE id = $1;

-- name: ListBranchSyncConflictsPage :many
SELECT
  c.id, c.branch_id, c.entity, c.ref, c.kind, c.branch_payload,
  c.central_client_id, c.status, c.resolution, c.resolved_by,
  c.created_at, c.updated_at,
  cl.name  AS central_name,
  cl.phone AS central_phone
FROM branch_sync_conflicts c
LEFT JOIN clients cl ON cl.id = c.central_client_id
WHERE ($1::bigint = 0 OR c.branch_id = $1)
  AND ($2::text = '' OR c.status = $2)
  AND ($3::text = '' OR c.entity = $3)
  AND ($4::text = '' OR c.ref ILIKE '%' || $4 || '%' OR c.branch_payload::text ILIKE '%' || $4 || '%')
ORDER BY
  CASE WHEN $5::text = 'created_at' AND $6::text = 'asc'  THEN c.created_at END ASC,
  CASE WHEN $5::text = 'created_at' AND $6::text = 'desc' THEN c.created_at END DESC,
  c.id DESC
LIMIT $7 OFFSET $8;

-- name: CountBranchSyncConflicts :one
SELECT COUNT(*) FROM branch_sync_conflicts
WHERE ($1::bigint = 0 OR branch_id = $1)
  AND ($2::text = '' OR status = $2)
  AND ($3::text = '' OR entity = $3)
  AND ($4::text = '' OR ref ILIKE '%' || $4 || '%' OR branch_payload::text ILIKE '%' || $4 || '%');

-- name: CountOpenBranchSyncConflicts :one
SELECT COUNT(*) FROM branch_sync_conflicts
WHERE branch_id = $1 AND status = 'open';

-- name: ResolveBranchSyncConflict :one
UPDATE branch_sync_conflicts
SET status = $2, resolution = $3, resolved_by = $4, updated_at = now()
WHERE id = $1
RETURNING *;
