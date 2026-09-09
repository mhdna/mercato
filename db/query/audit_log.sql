-- name: CreateAuditLog :one
INSERT INTO audit_logs (
  entity_type,
  entity_id,
  action,
  before,
  after,
  actor_id,
  source
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAuditLogs :many
SELECT
  al.*,
  COALESCE(u.name, '') AS actor_name
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_id
WHERE (sqlc.narg(entity_type)::text IS NULL OR al.entity_type = sqlc.narg(entity_type))
  AND (sqlc.narg(entity_id)::bigint IS NULL OR al.entity_id = sqlc.narg(entity_id))
  AND (sqlc.narg(action)::audit_action IS NULL OR al.action = sqlc.narg(action))
  AND (sqlc.narg(actor_id)::bigint IS NULL OR al.actor_id = sqlc.narg(actor_id))
  AND (sqlc.narg(from_date)::timestamptz IS NULL OR al.created_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::timestamptz IS NULL OR al.created_at <= sqlc.narg(to_date))
ORDER BY al.created_at DESC, al.id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountAuditLogs :one
SELECT COUNT(*)
FROM audit_logs al
WHERE (sqlc.narg(entity_type)::text IS NULL OR al.entity_type = sqlc.narg(entity_type))
  AND (sqlc.narg(entity_id)::bigint IS NULL OR al.entity_id = sqlc.narg(entity_id))
  AND (sqlc.narg(action)::audit_action IS NULL OR al.action = sqlc.narg(action))
  AND (sqlc.narg(actor_id)::bigint IS NULL OR al.actor_id = sqlc.narg(actor_id))
  AND (sqlc.narg(from_date)::timestamptz IS NULL OR al.created_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::timestamptz IS NULL OR al.created_at <= sqlc.narg(to_date));
