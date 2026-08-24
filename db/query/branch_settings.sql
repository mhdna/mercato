-- name: UpsertBranchSettings :one
INSERT INTO branch_settings (
  branch_id,
  branch_name,
  tax_rate,
  rounding_mode,
  rounding_currency,
  exchange_rate,
  exchange_window_hours,
  market_name,
  market_phone,
  market_description,
  return_policy,
  website,
  instagram,
  social_platforms,
  social_handles,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, now()
)
ON CONFLICT (branch_id) DO UPDATE SET
  branch_name = EXCLUDED.branch_name,
  tax_rate = EXCLUDED.tax_rate,
  rounding_mode = EXCLUDED.rounding_mode,
  rounding_currency = EXCLUDED.rounding_currency,
  exchange_rate = EXCLUDED.exchange_rate,
  exchange_window_hours = EXCLUDED.exchange_window_hours,
  market_name = EXCLUDED.market_name,
  market_phone = EXCLUDED.market_phone,
  market_description = EXCLUDED.market_description,
  return_policy = EXCLUDED.return_policy,
  website = EXCLUDED.website,
  instagram = EXCLUDED.instagram,
  social_platforms = EXCLUDED.social_platforms,
  social_handles = EXCLUDED.social_handles,
  updated_at = now()
RETURNING *;

-- name: GetBranchSettings :one
SELECT * FROM branch_settings WHERE branch_id = $1;
