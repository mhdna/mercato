-- Low stock alerts. A product's on-hand is the sum of inventory_stock
-- across every SKU (variant) and every inventory it's stocked in; its
-- effective threshold is its own override if set, else the single global
-- app_settings.default_low_stock_threshold.

-- name: ListLowStockProducts :many
-- status narrows to one bucket: 'out_of_stock' (on_hand <= 0) or
-- 'low_stock' (on_hand > 0, i.e. low but not yet out); NULL keeps both,
-- matching the "at or below threshold" definition used everywhere else.
WITH stock AS (
  SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
  FROM product_variants pv
  LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
  GROUP BY pv.product_id
)
SELECT
  p.id,
  p.code,
  p.name,
  p.low_stock_threshold,
  COALESCE(s.on_hand, 0)::bigint AS on_hand,
  COALESCE(p.low_stock_threshold, a.default_low_stock_threshold)::bigint AS effective_threshold
FROM products p
LEFT JOIN stock s ON s.product_id = p.id
CROSS JOIN app_settings a
WHERE p.is_active
  AND p.low_stock_alerts_enabled
  AND COALESCE(s.on_hand, 0) <= COALESCE(p.low_stock_threshold, a.default_low_stock_threshold)
  AND (sqlc.narg(status)::text IS NULL
       OR (sqlc.narg(status)::text = 'out_of_stock' AND COALESCE(s.on_hand, 0) <= 0)
       OR (sqlc.narg(status)::text = 'low_stock' AND COALESCE(s.on_hand, 0) > 0))
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%')
ORDER BY on_hand ASC, p.name
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountLowStockProducts :one
WITH stock AS (
  SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
  FROM product_variants pv
  LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
  GROUP BY pv.product_id
)
SELECT count(*)
FROM products p
LEFT JOIN stock s ON s.product_id = p.id
CROSS JOIN app_settings a
WHERE p.is_active
  AND p.low_stock_alerts_enabled
  AND COALESCE(s.on_hand, 0) <= COALESCE(p.low_stock_threshold, a.default_low_stock_threshold)
  AND (sqlc.narg(status)::text IS NULL
       OR (sqlc.narg(status)::text = 'out_of_stock' AND COALESCE(s.on_hand, 0) <= 0)
       OR (sqlc.narg(status)::text = 'low_stock' AND COALESCE(s.on_hand, 0) > 0))
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%');

-- name: GetLowStockSummary :one
-- Counts across all active, alert-enabled products, ignoring search --
-- backs the summary cards, which should reflect the whole picture even
-- while the table below is filtered.
WITH stock AS (
  SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
  FROM product_variants pv
  LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
  GROUP BY pv.product_id
)
SELECT
  count(*) FILTER (
    WHERE COALESCE(s.on_hand, 0) <= COALESCE(p.low_stock_threshold, a.default_low_stock_threshold)
      AND COALESCE(s.on_hand, 0) > 0
  )::bigint AS low_stock_count,
  count(*) FILTER (WHERE COALESCE(s.on_hand, 0) <= 0)::bigint AS out_of_stock_count,
  count(*)::bigint AS total_products
FROM products p
LEFT JOIN stock s ON s.product_id = p.id
CROSS JOIN app_settings a
WHERE p.is_active
  AND p.low_stock_alerts_enabled;

-- name: ListIgnoredLowStockProducts :many
-- Products that have opted out of low-stock alerting entirely, regardless
-- of on-hand or threshold -- lets the muted list be reviewed and reversed.
SELECT
  p.id,
  p.code,
  p.name,
  p.low_stock_threshold,
  COALESCE(s.on_hand, 0)::bigint AS on_hand,
  COALESCE(p.low_stock_threshold, a.default_low_stock_threshold)::bigint AS effective_threshold
FROM products p
LEFT JOIN (
  SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
  FROM product_variants pv
  LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
  GROUP BY pv.product_id
) s ON s.product_id = p.id
CROSS JOIN app_settings a
WHERE p.is_active
  AND NOT p.low_stock_alerts_enabled
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%')
ORDER BY p.name
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountIgnoredLowStockProducts :one
SELECT count(*)
FROM products p
WHERE p.is_active
  AND NOT p.low_stock_alerts_enabled
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%');

-- name: UpdateProductLowStockThreshold :exec
-- A null threshold clears the override, falling back to the app-wide default.
UPDATE products
SET low_stock_threshold = sqlc.narg(low_stock_threshold)
WHERE id = sqlc.arg(id);

-- name: SetProductLowStockAlertsEnabled :exec
UPDATE products
SET low_stock_alerts_enabled = $2
WHERE id = $1;

-- name: SetDefaultLowStockThreshold :one
UPDATE app_settings
SET default_low_stock_threshold = $1
WHERE id = 1
RETURNING *;
