-- Inventory health: overall days-of-inventory (on-hand / trailing daily
-- sell-through), total inventory value at moving-average cost, and the
-- same breakdown per inventory location.

-- name: GetStockHealthOverview :one
WITH sold AS (
  SELECT COALESCE(SUM(-quantity), 0)::bigint AS units_sold_30d
  FROM stock_movements
  WHERE reason = 'sale' AND created_at >= now() - interval '30 days'
),
stock AS (
  SELECT
    COALESCE(SUM(quantity), 0)::bigint AS total_units,
    COALESCE(SUM(quantity * avg_cost), 0)::bigint AS inventory_value
  FROM inventory_stock
)
SELECT
  stock.total_units,
  stock.inventory_value,
  sold.units_sold_30d,
  CASE WHEN sold.units_sold_30d > 0
    THEN ROUND(stock.total_units::numeric / (sold.units_sold_30d::numeric / 30), 1)::float8
    ELSE NULL
  END AS days_of_inventory
FROM stock CROSS JOIN sold;

-- name: ListStockHealthByInventory :many
WITH sold AS (
  SELECT inventory_id, COALESCE(SUM(-quantity), 0)::bigint AS units_sold_30d
  FROM stock_movements
  WHERE reason = 'sale' AND created_at >= now() - interval '30 days'
  GROUP BY inventory_id
),
stock AS (
  SELECT
    inventory_id,
    COALESCE(SUM(quantity), 0)::bigint AS total_units,
    COALESCE(SUM(quantity * avg_cost), 0)::bigint AS inventory_value
  FROM inventory_stock
  GROUP BY inventory_id
)
SELECT
  i.id,
  i.name,
  i.type,
  COALESCE(stock.total_units, 0)::bigint AS total_units,
  COALESCE(stock.inventory_value, 0)::bigint AS inventory_value,
  COALESCE(sold.units_sold_30d, 0)::bigint AS units_sold_30d,
  CASE WHEN COALESCE(sold.units_sold_30d, 0) > 0
    THEN ROUND(COALESCE(stock.total_units, 0)::numeric / (sold.units_sold_30d::numeric / 30), 1)::float8
    ELSE NULL
  END AS days_of_inventory
FROM inventories i
LEFT JOIN stock ON stock.inventory_id = i.id
LEFT JOIN sold ON sold.inventory_id = i.id
WHERE (sqlc.arg(search)::text = '' OR i.name ILIKE '%' || sqlc.arg(search) || '%')
ORDER BY i.name
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountStockHealthByInventory :one
SELECT count(*)
FROM inventories i
WHERE (sqlc.arg(search)::text = '' OR i.name ILIKE '%' || sqlc.arg(search) || '%');

-- ---------------------------------------------------------------------------
-- Slow movers: money tied up in stock that isn't selling. Deliberately
-- separate from the low-stock alerts (too little stock) -- this is the
-- opposite problem, too much. A product qualifies when it's carrying stock
-- (on_hand > 0) and either never sold in the last 90 days ("dead": put it
-- on clearance) or would take unreasonably long to sell through at its
-- current pace ("overstocked": more than 90 days on hand).
-- ---------------------------------------------------------------------------

-- name: ListSlowMovingProducts :many
WITH velocity AS (
  SELECT
    p.id, p.code, p.name,
    COALESCE(stock.on_hand, 0)::bigint AS on_hand,
    COALESCE(sold.units_sold_90d, 0)::bigint AS units_sold_90d
  FROM products p
  LEFT JOIN (
    SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
    FROM product_variants pv
    LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
    GROUP BY pv.product_id
  ) stock ON stock.product_id = p.id
  LEFT JOIN (
    SELECT pv.product_id, COALESCE(SUM(-sm.quantity), 0)::bigint AS units_sold_90d
    FROM product_variants pv
    JOIN stock_movements sm ON sm.variant_id = pv.id
      AND sm.reason = 'sale' AND sm.created_at >= now() - interval '90 days'
    GROUP BY pv.product_id
  ) sold ON sold.product_id = p.id
  WHERE p.is_active AND COALESCE(stock.on_hand, 0) > 0
)
SELECT
  id, code, name, on_hand, units_sold_90d,
  CASE WHEN units_sold_90d > 0
    THEN ROUND(on_hand::numeric / (units_sold_90d::numeric / 90), 1)::float8
    ELSE NULL
  END AS days_of_inventory,
  CASE WHEN units_sold_90d = 0 THEN 'dead' ELSE 'overstocked' END AS category
FROM velocity
WHERE (units_sold_90d = 0 OR (on_hand::numeric / (units_sold_90d::numeric / 90)) > 90)
  AND (sqlc.narg(category)::text IS NULL
       OR (sqlc.narg(category)::text = 'dead' AND units_sold_90d = 0)
       OR (sqlc.narg(category)::text = 'overstocked' AND units_sold_90d > 0))
  AND (sqlc.arg(search)::text = ''
       OR name ILIKE '%' || sqlc.arg(search) || '%'
       OR code ILIKE '%' || sqlc.arg(search) || '%')
ORDER BY (units_sold_90d = 0) DESC, on_hand DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSlowMovingProducts :one
WITH velocity AS (
  SELECT
    p.id,
    COALESCE(stock.on_hand, 0)::bigint AS on_hand,
    COALESCE(sold.units_sold_90d, 0)::bigint AS units_sold_90d
  FROM products p
  LEFT JOIN (
    SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
    FROM product_variants pv
    LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
    GROUP BY pv.product_id
  ) stock ON stock.product_id = p.id
  LEFT JOIN (
    SELECT pv.product_id, COALESCE(SUM(-sm.quantity), 0)::bigint AS units_sold_90d
    FROM product_variants pv
    JOIN stock_movements sm ON sm.variant_id = pv.id
      AND sm.reason = 'sale' AND sm.created_at >= now() - interval '90 days'
    GROUP BY pv.product_id
  ) sold ON sold.product_id = p.id
  WHERE p.is_active AND COALESCE(stock.on_hand, 0) > 0
)
SELECT count(*)
FROM velocity v
JOIN products p ON p.id = v.id
WHERE (v.units_sold_90d = 0 OR (v.on_hand::numeric / (v.units_sold_90d::numeric / 90)) > 90)
  AND (sqlc.narg(category)::text IS NULL
       OR (sqlc.narg(category)::text = 'dead' AND v.units_sold_90d = 0)
       OR (sqlc.narg(category)::text = 'overstocked' AND v.units_sold_90d > 0))
  AND (sqlc.arg(search)::text = ''
       OR p.name ILIKE '%' || sqlc.arg(search) || '%'
       OR p.code ILIKE '%' || sqlc.arg(search) || '%');

-- name: GetSlowMoverSummary :one
WITH velocity AS (
  SELECT
    p.id,
    COALESCE(stock.on_hand, 0)::bigint AS on_hand,
    COALESCE(sold.units_sold_90d, 0)::bigint AS units_sold_90d
  FROM products p
  LEFT JOIN (
    SELECT pv.product_id, COALESCE(SUM(ist.quantity), 0)::bigint AS on_hand
    FROM product_variants pv
    LEFT JOIN inventory_stock ist ON ist.variant_id = pv.id
    GROUP BY pv.product_id
  ) stock ON stock.product_id = p.id
  LEFT JOIN (
    SELECT pv.product_id, COALESCE(SUM(-sm.quantity), 0)::bigint AS units_sold_90d
    FROM product_variants pv
    JOIN stock_movements sm ON sm.variant_id = pv.id
      AND sm.reason = 'sale' AND sm.created_at >= now() - interval '90 days'
    GROUP BY pv.product_id
  ) sold ON sold.product_id = p.id
  WHERE p.is_active
)
SELECT
  count(*) FILTER (WHERE on_hand > 0 AND units_sold_90d = 0)::bigint AS dead_stock_count,
  count(*) FILTER (
    WHERE on_hand > 0 AND units_sold_90d > 0
      AND (on_hand::numeric / (units_sold_90d::numeric / 90)) > 90
  )::bigint AS overstocked_count,
  count(*)::bigint AS total_active_products
FROM velocity;
