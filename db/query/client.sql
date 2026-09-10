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

-- name: ListClientsWithLoyalty :many
-- Same rows as ListClients, but with each branch's reported loyalty points
-- (branch_invoices.loyalty_points_delta, via client_links) summed in
-- alongside kashi's own admin-invoice points. Used by the admin clients
-- page so the displayed balance isn't missing everything earned at a
-- branch till.
SELECT c.*, COALESCE(bl.branch_points, 0)::bigint AS branch_loyalty_points
FROM clients c
LEFT JOIN (
  SELECT cl.client_id, SUM(bi.loyalty_points_delta) AS branch_points
  FROM branch_invoices bi
  JOIN client_links cl ON cl.branch_id = bi.branch_id AND cl.branch_client_id = bi.branch_client_id
  GROUP BY cl.client_id
) bl ON bl.client_id = c.id
WHERE sqlc.narg(client_type)::text IS NULL OR c.client_type = sqlc.narg(client_type)
ORDER BY c.name
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

-- name: ListClientsBySpending :many
-- Powers the clients "loyalty pyramid" + rest-of-clients table: every
-- client ranked by total_spent, invoice_count, or item_count (chosen via
-- sort_by -- see the ORDER BY at the bottom). Two sources feed each metric:
-- kashi's own admin invoices (sales positive, returns negative --
-- invoices.grand_total is always stored as a positive magnitude regardless
-- of kind, see ReturnInvoiceTx) and branch-reported sales
-- (branch_invoices.grand_total, which already carries the correct net sign
-- -- see ListDailyIncome). sqlc.narg(branch_id) scopes to one branch's
-- till; central admin invoices aren't attributed to any branch, so they're
-- excluded entirely once a single branch is selected (an "All Branches"
-- view is the only one where they belong). invoice_count/item_count follow
-- the same all-or-one-branch rule for consistency with total_spent.
-- invoice_count/item_count are computed in the `ranked` CTE rather than as
-- plain SELECT-list aliases so the ORDER BY's CASE expression can actually
-- reference them: Postgres only resolves an output alias in ORDER BY when
-- it's used bare, not when it's embedded in a larger expression (there it's
-- looked up as an ordinary column instead, which fails since it isn't one
-- of the FROM clause's columns). Wrapping in a CTE turns them into real
-- columns of the derived table, so the CASE inside ORDER BY resolves fine.
WITH ranked AS (
  SELECT
    c.*,
    (
      CASE WHEN sqlc.narg(branch_id)::bigint IS NULL THEN COALESCE(ci.central_total, 0) ELSE 0 END
      + COALESCE(bi.branch_total, 0)
    )::bigint AS total_spent,
    (
      CASE WHEN sqlc.narg(branch_id)::bigint IS NULL THEN COALESCE(ci.central_invoice_count, 0) ELSE 0 END
      + COALESCE(bi.branch_invoice_count, 0)
    )::bigint AS invoice_count,
    (
      CASE WHEN sqlc.narg(branch_id)::bigint IS NULL THEN COALESCE(ci.central_item_count, 0) ELSE 0 END
      + COALESCE(bi.branch_item_count, 0)
    )::bigint AS item_count
  FROM clients c
  LEFT JOIN (
    SELECT i.client_id,
      SUM(CASE WHEN si.invoice_id IS NOT NULL THEN i.grand_total ELSE -i.grand_total END) AS central_total,
      COUNT(DISTINCT i.id) AS central_invoice_count,
      COALESCE(SUM(ip.quantity), 0) AS central_item_count
    FROM invoices i
    LEFT JOIN sales_invoices si ON si.invoice_id = i.id
    LEFT JOIN invoice_products ip ON ip.invoice_id = i.id
    GROUP BY i.client_id
  ) ci ON ci.client_id = c.id
  LEFT JOIN (
    SELECT cl.client_id,
      SUM(b.grand_total) AS branch_total,
      COUNT(DISTINCT b.id) AS branch_invoice_count,
      COALESCE(SUM(bii.quantity), 0) AS branch_item_count
    FROM branch_invoices b
    JOIN client_links cl ON cl.branch_id = b.branch_id AND cl.branch_client_id = b.branch_client_id
    LEFT JOIN branch_invoice_items bii ON bii.branch_invoice_id = b.id
    WHERE sqlc.narg(branch_id)::bigint IS NULL OR b.branch_id = sqlc.narg(branch_id)
    GROUP BY cl.client_id
  ) bi ON bi.client_id = c.id
  WHERE (sqlc.narg(client_type)::text IS NULL OR c.client_type = sqlc.narg(client_type))
    AND (sqlc.narg(search)::text IS NULL OR c.name ILIKE '%' || sqlc.narg(search)::text || '%' OR c.phone ILIKE '%' || sqlc.narg(search)::text || '%')
)
SELECT *
FROM ranked
ORDER BY
  (CASE sqlc.arg(sort_by)::text
    WHEN 'invoices' THEN invoice_count
    WHEN 'items' THEN item_count
    ELSE total_spent
  END) DESC,
  name ASC
LIMIT $1
OFFSET $2;

-- name: CountClientsBySpending :one
-- Same filters as ListClientsBySpending, minus the spend computation --
-- used for the rest-of-clients table's pagination total. Deliberately
-- ignores branch_id: the count of matching clients doesn't change with
-- which branch's sales are being summed, only who qualifies by type/search.
SELECT COUNT(*) FROM clients c
WHERE (sqlc.narg(client_type)::text IS NULL OR c.client_type = sqlc.narg(client_type))
  AND (sqlc.narg(search)::text IS NULL OR c.name ILIKE '%' || sqlc.narg(search)::text || '%' OR c.phone ILIKE '%' || sqlc.narg(search)::text || '%');

-- name: ListClientsUpdatedSincePaged :many
-- Keyset-paginated form of ListClientsUpdatedSince for a branch doing a
-- full catch-up (cursor reset): ordered by (updated_at, id) so a page
-- boundary that falls between two rows sharing an updated_at can't drop or
-- repeat one. after_updated_at/after_id are the last row of the previous
-- page (zero-time / 0 for the first page).
SELECT * FROM clients
WHERE client_type = 'retail'
  AND (updated_at, id) > (sqlc.arg(after_updated_at)::timestamptz, sqlc.arg(after_id)::bigint)
ORDER BY updated_at, id
LIMIT sqlc.arg(row_limit);
