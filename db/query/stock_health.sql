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
