-- Queries behind the home dashboard's Overview tab: the summary cards, the
-- revenue/expense trend charts, and the aggregated Sales / Purchases /
-- Recent Expenses tables. Every query takes a half-open [date_from,
-- date_to) timestamptz range (the handler passes start-of-next-day as
-- date_to) and, where a branch dimension exists, a nullable branch_id
-- (NULL = all branches, same convention as ListBranchInvoices).
--
-- "branch" data is what the POS branches sync up (branch_invoices,
-- branch_expenses); "admin" data is kashi's own centrally-entered records
-- (invoices/sales_invoices, expenses). The dashboard's scope filter picks
-- one or both; the list queries take that as an explicit `scope` arg so a
-- single UNION ALL can serve all three modes, while the scalar summary
-- queries are split per source and the handler adds up whichever apply.

-- name: DashboardActivitiesList :many
-- One chronological feed across every kind of persisted branch sync. The
-- union happens before LIMIT/OFFSET so pagination is correct globally, not
-- independently per source table. received_at is used deliberately: this
-- table answers "what synced recently", even when a branch catches up an
-- older occurred_at record.
WITH activities AS (
  SELECT
    'Invoice'::text AS activity,
    b.branch_id,
    br.name AS branch_name,
    concat(initcap(b.kind), ' · ', b.branch_invoice_code)::text AS details,
    b.grand_total::bigint AS amount,
    ''::text AS currency_code,
    b.received_at AS synced_at
  FROM branch_invoices b
  JOIN branches br ON br.id = b.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND b.received_at >= sqlc.arg(date_from)::timestamptz
    AND b.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR b.branch_id = sqlc.narg(branch_id))

  UNION ALL
  SELECT 'Expense', e.branch_id, br.name, e.description,
    -e.amount, e.currency_code, e.received_at
  FROM branch_expenses e
  JOIN branches br ON br.id = e.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND e.received_at >= sqlc.arg(date_from)::timestamptz
    AND e.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR e.branch_id = sqlc.narg(branch_id))

  UNION ALL
  SELECT 'Loan', l.branch_id, br.name, l.description,
    l.amount, l.currency_code, l.received_at
  FROM loans l
  JOIN branches br ON br.id = l.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND l.origin = 'branch_loan'
    AND l.received_at >= sqlc.arg(date_from)::timestamptz
    AND l.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR l.branch_id = sqlc.narg(branch_id))

  UNION ALL
  SELECT 'Shift closed', s.branch_id, br.name,
    concat_ws(' · ', nullif(s.closing_person_name, ''), 'USD variance')::text,
    s.variance_usd, 'USD', s.received_at
  FROM branch_shifts s
  JOIN branches br ON br.id = s.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND s.received_at >= sqlc.arg(date_from)::timestamptz
    AND s.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR s.branch_id = sqlc.narg(branch_id))

  UNION ALL
  SELECT 'Settlement changed', s.branch_id, br.name,
    concat('Invoice ', s.sale_client_ref)::text,
    s.grand_total, '', s.received_at
  FROM branch_invoice_settlements s
  JOIN branches br ON br.id = s.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND s.received_at >= sqlc.arg(date_from)::timestamptz
    AND s.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR s.branch_id = sqlc.narg(branch_id))

  UNION ALL
  SELECT 'Attendance', a.branch_id, br.name,
    concat_ws(' · ', a.salesperson_name, a.type, a.event_date, a.event_time)::text,
    0::bigint, '', a.received_at
  FROM branch_attendance_events a
  JOIN branches br ON br.id = a.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND a.received_at >= sqlc.arg(date_from)::timestamptz
    AND a.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR a.branch_id = sqlc.narg(branch_id))

  UNION ALL
  SELECT 'Attendance changed', a.branch_id, br.name,
    concat_ws(' · ', a.salesperson_name, a.kind, a.status, a.note)::text,
    0::bigint, '', a.received_at
  FROM branch_attendance_changes a
  JOIN branches br ON br.id = a.branch_id
  WHERE sqlc.arg(scope)::text <> 'admin'
    AND a.received_at >= sqlc.arg(date_from)::timestamptz
    AND a.received_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR a.branch_id = sqlc.narg(branch_id))
)
SELECT activities.*, COUNT(*) OVER()::bigint AS total_count
FROM activities
ORDER BY synced_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_id);

-- name: DashboardBranchSummary :one
-- items_sold / revenue / invoice count for branch sales in the range.
-- Returns are excluded from items_sold and the count (they aren't sales),
-- but revenue sums every kind so refunds net out, matching ListDailyIncome.
SELECT
  COALESCE((
    SELECT SUM(bii.quantity)
    FROM branch_invoice_items bii
    JOIN branch_invoices b2 ON b2.id = bii.branch_invoice_id
    WHERE b2.kind = 'sales'
      AND b2.occurred_at >= sqlc.arg(date_from)::timestamptz
      AND b2.occurred_at < sqlc.arg(date_to)::timestamptz
      AND (sqlc.narg(branch_id)::bigint IS NULL OR b2.branch_id = sqlc.narg(branch_id))
  ), 0)::bigint AS items_sold,
  COALESCE(SUM(b.grand_total), 0)::bigint AS revenue,
  COUNT(*) FILTER (WHERE b.kind = 'sales')::bigint AS invoices
FROM branch_invoices b
WHERE b.occurred_at >= sqlc.arg(date_from)::timestamptz
  AND b.occurred_at < sqlc.arg(date_to)::timestamptz
  AND (sqlc.narg(branch_id)::bigint IS NULL OR b.branch_id = sqlc.narg(branch_id));

-- name: DashboardBranchExpensesTotal :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total
FROM branch_expenses
WHERE occurred_at >= sqlc.arg(date_from)::timestamptz
  AND occurred_at < sqlc.arg(date_to)::timestamptz
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id));

-- name: DashboardBranchNewClients :one
-- Clients first seen at a branch in the range. With a branch selected,
-- count distinct linked clients whose central row was created in-range;
-- with no branch, every client created in-range.
SELECT COUNT(*)::bigint AS total
FROM clients c
WHERE c.created_at >= sqlc.arg(date_from)::timestamptz
  AND c.created_at < sqlc.arg(date_to)::timestamptz
  AND (
    sqlc.narg(branch_id)::bigint IS NULL
    OR EXISTS (
      SELECT 1 FROM client_links cl
      WHERE cl.client_id = c.id AND cl.branch_id = sqlc.narg(branch_id)
    )
  );

-- name: DashboardAdminSummary :one
SELECT
  COALESCE((
    SELECT SUM(ip.quantity)
    FROM invoice_products ip
    JOIN sales_invoices si ON si.invoice_id = ip.invoice_id
    JOIN invoices i2 ON i2.id = ip.invoice_id
    WHERE i2.created_at >= sqlc.arg(date_from)::timestamptz
      AND i2.created_at < sqlc.arg(date_to)::timestamptz
  ), 0)::bigint AS items_sold,
  COALESCE(SUM(i.grand_total), 0)::bigint AS revenue,
  COUNT(*)::bigint AS invoices
FROM invoices i
JOIN sales_invoices si ON si.invoice_id = i.id
WHERE i.created_at >= sqlc.arg(date_from)::timestamptz
  AND i.created_at < sqlc.arg(date_to)::timestamptz;

-- name: DashboardAdminExpensesTotal :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total
FROM expenses
WHERE created_at >= sqlc.arg(date_from)::timestamptz
  AND created_at < sqlc.arg(date_to)::timestamptz;

-- name: DashboardAdminNewClients :one
SELECT COUNT(*)::bigint AS total
FROM clients
WHERE created_at >= sqlc.arg(date_from)::timestamptz
  AND created_at < sqlc.arg(date_to)::timestamptz;

-- name: DashboardDueLoans :one
-- Outstanding loan balance = disbursed - repaid. loans has no due date, so
-- "due" here just means still owed. Branch arm = origin 'branch_loan'
-- (scoped by branch_id when set); admin arm = origin 'central_loan'.
-- loan_payments carry no branch dimension of their own, so they follow
-- their parent loan's arm. Not date-range bound -- a balance, not a flow.
SELECT (
  COALESCE((
    SELECT SUM(l.amount) FROM loans l
    WHERE (
      (sqlc.arg(scope)::text IN ('all', 'branch') AND l.origin = 'branch_loan'
        AND (sqlc.narg(branch_id)::bigint IS NULL OR l.branch_id = sqlc.narg(branch_id)))
      OR (sqlc.arg(scope)::text IN ('all', 'admin') AND l.origin = 'central_loan')
    )
  ), 0)
  -
  COALESCE((
    SELECT SUM(lp.amount) FROM loan_payments lp
    JOIN loans l ON l.id = lp.loan_id
    WHERE (
      (sqlc.arg(scope)::text IN ('all', 'branch') AND l.origin = 'branch_loan'
        AND (sqlc.narg(branch_id)::bigint IS NULL OR l.branch_id = sqlc.narg(branch_id)))
      OR (sqlc.arg(scope)::text IN ('all', 'admin') AND l.origin = 'central_loan')
    )
  ), 0)
)::bigint AS outstanding;

-- name: DashboardTargetList :many
-- Every branch_target that is active and spans `as_of` (today), with its
-- live revenue (the target's own [date_from, date_to] window against its
-- own branch). One progress bar per row on the dashboard; the handler also
-- sums these for the overall attainment percentage. branch_targets are
-- branch-only, so the handler skips this under admin scope.
SELECT
  t.id,
  t.branch_id,
  b.name AS branch_name,
  t.date_from,
  t.date_to,
  t.target_amount,
  t.color,
  COALESCE((
    SELECT SUM(bi.grand_total)
    FROM branch_invoices bi
    WHERE bi.branch_id = t.branch_id
      AND (bi.occurred_at AT TIME ZONE 'UTC')::date BETWEEN t.date_from AND t.date_to
  ), 0)::bigint AS revenue
FROM branch_targets t
JOIN branches b ON b.id = t.branch_id
WHERE t.is_active
  AND t.date_from <= sqlc.arg(as_of)::date
  AND t.date_to >= sqlc.arg(as_of)::date
  AND (sqlc.narg(branch_id)::bigint IS NULL OR t.branch_id = sqlc.narg(branch_id))
ORDER BY b.name, t.date_from;

-- name: DashboardTopDueLoans :many
-- The loans with the largest outstanding balance, for the dashboard's
-- "Due Loans" list. Same scope/branch gating as DashboardDueLoans; fully
-- repaid loans (outstanding <= 0) are dropped.
SELECT
  l.id,
  l.description,
  l.origin,
  COALESCE(b.name, '') AS branch_name,
  l.amount,
  (l.amount - COALESCE((
    SELECT SUM(lp.amount) FROM loan_payments lp WHERE lp.loan_id = l.id
  ), 0))::bigint AS outstanding
FROM loans l
LEFT JOIN branches b ON b.id = l.branch_id
WHERE (
    (sqlc.arg(scope)::text IN ('all', 'branch') AND l.origin = 'branch_loan'
      AND (sqlc.narg(branch_id)::bigint IS NULL OR l.branch_id = sqlc.narg(branch_id)))
    OR (sqlc.arg(scope)::text IN ('all', 'admin') AND l.origin = 'central_loan')
  )
  AND (l.amount - COALESCE((
    SELECT SUM(lp.amount) FROM loan_payments lp WHERE lp.loan_id = l.id
  ), 0)) > 0
ORDER BY outstanding DESC
LIMIT sqlc.arg(top_n)::int;

-- name: DashboardBranchRevenueSeries :many
SELECT
  (occurred_at AT TIME ZONE 'UTC')::date AS day,
  COALESCE(SUM(grand_total), 0)::bigint AS total
FROM branch_invoices
WHERE occurred_at >= sqlc.arg(date_from)::timestamptz
  AND occurred_at < sqlc.arg(date_to)::timestamptz
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
GROUP BY day
ORDER BY day;

-- name: DashboardBranchExpenseSeries :many
SELECT
  (occurred_at AT TIME ZONE 'UTC')::date AS day,
  COALESCE(SUM(amount), 0)::bigint AS total
FROM branch_expenses
WHERE occurred_at >= sqlc.arg(date_from)::timestamptz
  AND occurred_at < sqlc.arg(date_to)::timestamptz
  AND (sqlc.narg(branch_id)::bigint IS NULL OR branch_id = sqlc.narg(branch_id))
GROUP BY day
ORDER BY day;

-- name: DashboardAdminRevenueSeries :many
SELECT
  (i.created_at AT TIME ZONE 'UTC')::date AS day,
  COALESCE(SUM(i.grand_total), 0)::bigint AS total
FROM invoices i
JOIN sales_invoices si ON si.invoice_id = i.id
WHERE i.created_at >= sqlc.arg(date_from)::timestamptz
  AND i.created_at < sqlc.arg(date_to)::timestamptz
GROUP BY day
ORDER BY day;

-- name: DashboardAdminExpenseSeries :many
SELECT
  (created_at AT TIME ZONE 'UTC')::date AS day,
  COALESCE(SUM(amount), 0)::bigint AS total
FROM expenses
WHERE created_at >= sqlc.arg(date_from)::timestamptz
  AND created_at < sqlc.arg(date_to)::timestamptz
GROUP BY day
ORDER BY day;

-- name: DashboardSalesList :many
-- currency is the app's single default currency code (every amount in the
-- schema is stored in its minor unit -- see 000009_create_currencies).
-- The branch-invoice client name is resolved through client_links; when
-- the branch never synced a client link it comes back empty.
SELECT
  source::text AS source,
  client::text AS client,
  total::bigint AS total,
  (SELECT code FROM currencies WHERE is_default LIMIT 1)::text AS currency,
  occurred_at::timestamptz AS occurred_at
FROM (
  SELECT
    'branch'::text AS source,
    COALESCE(bc.name, '')::text AS client,
    b.grand_total AS total,
    b.occurred_at AS occurred_at
  FROM branch_invoices b
  LEFT JOIN client_links cl
    ON cl.branch_id = b.branch_id AND cl.branch_client_id = b.branch_client_id
  LEFT JOIN clients bc ON bc.id = cl.client_id
  WHERE b.kind = 'sales'
    AND sqlc.arg(scope)::text IN ('all', 'branch')
    AND b.occurred_at >= sqlc.arg(date_from)::timestamptz
    AND b.occurred_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR b.branch_id = sqlc.narg(branch_id))

  UNION ALL

  SELECT
    'admin'::text AS source,
    COALESCE(c.name, '')::text AS client,
    i.grand_total AS total,
    i.created_at AS occurred_at
  FROM invoices i
  JOIN sales_invoices si ON si.invoice_id = i.id
  LEFT JOIN clients c ON c.id = i.client_id
  WHERE sqlc.arg(scope)::text IN ('all', 'admin')
    AND i.created_at >= sqlc.arg(date_from)::timestamptz
    AND i.created_at < sqlc.arg(date_to)::timestamptz
) rows
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDashboardSales :one
SELECT (
  (
    SELECT COUNT(*) FROM branch_invoices b
    WHERE b.kind = 'sales'
      AND sqlc.arg(scope)::text IN ('all', 'branch')
      AND b.occurred_at >= sqlc.arg(date_from)::timestamptz
      AND b.occurred_at < sqlc.arg(date_to)::timestamptz
      AND (sqlc.narg(branch_id)::bigint IS NULL OR b.branch_id = sqlc.narg(branch_id))
  )
  +
  (
    SELECT COUNT(*) FROM invoices i
    JOIN sales_invoices si ON si.invoice_id = i.id
    WHERE sqlc.arg(scope)::text IN ('all', 'admin')
      AND i.created_at >= sqlc.arg(date_from)::timestamptz
      AND i.created_at < sqlc.arg(date_to)::timestamptz
  )
)::bigint AS total;

-- name: DashboardPurchasesList :many
-- Purchases are admin-only -- there is no branch purchase concept -- so the
-- scope arg only ever excludes them (scope = 'branch'), never adds a second
-- source. total is summed from purchase_items since purchases carries no
-- amount column of its own.
SELECT
  p.id AS id,
  s.name AS supplier,
  COALESCE((
    SELECT SUM(pi.quantity * pi.unit_price)
    FROM purchase_items pi
    WHERE pi.purchase_id = p.id
  ), 0)::bigint AS total,
  COALESCE((
    SELECT SUM(pi.quantity)
    FROM purchase_items pi
    WHERE pi.purchase_id = p.id
  ), 0)::bigint AS items,
  (SELECT code FROM currencies WHERE is_default LIMIT 1)::text AS currency,
  p.purchased_at AS purchased_at
FROM purchases p
JOIN suppliers s ON s.id = p.supplier_id
WHERE sqlc.arg(scope)::text IN ('all', 'admin')
  AND p.purchased_at >= sqlc.arg(date_from)::timestamptz
  AND p.purchased_at < sqlc.arg(date_to)::timestamptz
ORDER BY p.purchased_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDashboardPurchases :one
SELECT COUNT(*)::bigint AS total
FROM purchases p
WHERE sqlc.arg(scope)::text IN ('all', 'admin')
  AND p.purchased_at >= sqlc.arg(date_from)::timestamptz
  AND p.purchased_at < sqlc.arg(date_to)::timestamptz;

-- name: DashboardPurchasesTotal :one
-- Cost of goods bought in the range. Admin-only (no branch purchase
-- concept); total is summed from purchase_items since purchases carries no
-- amount column of its own, matching DashboardPurchasesList.
SELECT COALESCE(SUM(pi.quantity * pi.unit_price), 0)::bigint AS total
FROM purchases p
JOIN purchase_items pi ON pi.purchase_id = p.id
WHERE p.purchased_at >= sqlc.arg(date_from)::timestamptz
  AND p.purchased_at < sqlc.arg(date_to)::timestamptz;

-- name: DashboardPurchaseSeries :many
SELECT
  (p.purchased_at AT TIME ZONE 'UTC')::date AS day,
  COALESCE(SUM(pi.quantity * pi.unit_price), 0)::bigint AS total
FROM purchases p
JOIN purchase_items pi ON pi.purchase_id = p.id
WHERE p.purchased_at >= sqlc.arg(date_from)::timestamptz
  AND p.purchased_at < sqlc.arg(date_to)::timestamptz
GROUP BY day
ORDER BY day;

-- name: DashboardExpensesList :many
SELECT
  source::text AS source,
  description::text AS description,
  category::text AS category,
  amount::bigint AS amount,
  (SELECT code FROM currencies WHERE is_default LIMIT 1)::text AS currency,
  occurred_at::timestamptz AS occurred_at,
  image_id::bigint AS image_id
-- image_id is 0, not NULL, when an expense has no receipt photo (admin
-- expenses never do): keeps the generated column a plain int64 and lets
-- the frontend treat 0 as "no image".
FROM (
  SELECT
    'branch'::text AS source,
    be.description AS description,
    COALESCE(ec.name, '')::text AS category,
    be.amount AS amount,
    be.occurred_at AS occurred_at,
    COALESCE((
      SELECT bei.id FROM branch_expense_images bei
      WHERE bei.branch_expense_id = be.id
      ORDER BY bei.id
      LIMIT 1
    ), 0) AS image_id
  FROM branch_expenses be
  LEFT JOIN expense_categories ec ON ec.id = be.category_id
  WHERE sqlc.arg(scope)::text IN ('all', 'branch')
    AND be.occurred_at >= sqlc.arg(date_from)::timestamptz
    AND be.occurred_at < sqlc.arg(date_to)::timestamptz
    AND (sqlc.narg(branch_id)::bigint IS NULL OR be.branch_id = sqlc.narg(branch_id))

  UNION ALL

  SELECT
    'admin'::text AS source,
    e.description AS description,
    COALESCE(ec.name, '')::text AS category,
    e.amount AS amount,
    e.created_at AS occurred_at,
    0::bigint AS image_id
  FROM expenses e
  LEFT JOIN expense_categories ec ON ec.id = e.category_id
  WHERE sqlc.arg(scope)::text IN ('all', 'admin')
    AND e.created_at >= sqlc.arg(date_from)::timestamptz
    AND e.created_at < sqlc.arg(date_to)::timestamptz
) rows
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDashboardExpenses :one
SELECT (
  (
    SELECT COUNT(*) FROM branch_expenses be
    WHERE sqlc.arg(scope)::text IN ('all', 'branch')
      AND be.occurred_at >= sqlc.arg(date_from)::timestamptz
      AND be.occurred_at < sqlc.arg(date_to)::timestamptz
      AND (sqlc.narg(branch_id)::bigint IS NULL OR be.branch_id = sqlc.narg(branch_id))
  )
  +
  (
    SELECT COUNT(*) FROM expenses e
    WHERE sqlc.arg(scope)::text IN ('all', 'admin')
      AND e.created_at >= sqlc.arg(date_from)::timestamptz
      AND e.created_at < sqlc.arg(date_to)::timestamptz
  )
)::bigint AS total;
