-- name: CreateClient :one
INSERT INTO clients (
  name,
  phone,
  client_type
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetClient :one
SELECT * FROM clients
WHERE id = $1 LIMIT 1;

-- name: ListClients :many
-- sqlc.narg(client_type) is nullable: NULL means "all types" -- the admin
-- clients page passes it when a Retail/Wholesale tab is selected, and
-- omits it for a combined view if one is ever added.
SELECT * FROM clients
WHERE sqlc.narg(client_type)::text IS NULL OR client_type = sqlc.narg(client_type)
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: UpdateClient :one
UPDATE clients
  SET name = $2,
  phone = $3,
  client_type = $4,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetClientByPhone :one
SELECT * FROM clients
WHERE phone = $1 LIMIT 1;

-- name: ListClientsUpdatedSince :many
-- Feeds the branch catch-up endpoint (see branch_catchup.go), same pattern
-- as currencies/cashbox_accounts/products. Wholesale clients never appear
-- here -- they're a central-office concept, not something a branch till
-- should ever see in its client search. Converting an already-synced
-- retail client to wholesale won't retract it from branches that already
-- have it locally (no delete propagation exists for any synced entity
-- today); acceptable since that conversion is expected to be rare.
SELECT * FROM clients
WHERE updated_at > $1 AND client_type = 'retail'
ORDER BY updated_at;

-- name: UpsertClientLink :exec
INSERT INTO client_links (branch_id, branch_client_id, client_id)
VALUES ($1, $2, $3)
ON CONFLICT (branch_id, branch_client_id) DO UPDATE SET client_id = EXCLUDED.client_id;

-- name: GetClientLink :one
SELECT client_id FROM client_links
WHERE branch_id = $1 AND branch_client_id = $2;

-- name: SumClientLoyaltyPoints :one
-- The redemption-time query: every branch's contribution to this client's
-- loyalty balance, added up. Each branch is decisive for what it reports
-- (loyalty_points_delta on branch_invoices) -- this never recomputes that
-- number, only sums what's already been reported.
SELECT COALESCE(SUM(bi.loyalty_points_delta), 0)::bigint AS total
FROM branch_invoices bi
JOIN client_links cl ON cl.branch_id = bi.branch_id AND cl.branch_client_id = bi.branch_client_id
WHERE cl.client_id = $1;

-- name: DeleteClient :exec
DELETE FROM clients
WHERE id = $1;

-- name: AddClientLoyaltyPoints :exec
UPDATE clients 
  SET total_loyalty_points = total_loyalty_points + $2,
  valid_loyalty_points = valid_loyalty_points + $3
WHERE id = $1;
-- name: CountClients :one
SELECT COUNT(*) FROM clients
WHERE sqlc.narg(client_type)::text IS NULL OR client_type = sqlc.narg(client_type);
