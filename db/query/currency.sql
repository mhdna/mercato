-- name: CreateCurrency :one
INSERT INTO currencies (
  name,
  code,
  symbol,
  value_in_default_currency,
  units_per_usd_micros,
  cash_rounding_unit
)
VALUES ( $1, $2, $3, $4, $5, $6 )
RETURNING *;

-- name: DeleteCurrency :exec
delete from currencies
where code = $1;

-- name: DeleteCurrencies :execrows
delete from currencies
where code = ANY(sqlc.arg(codes)::text[]);

-- name: UpdateCurrency :one
UPDATE currencies
SET name = $2,
symbol = $3,
value_in_default_currency = $4,
is_active = $5,
units_per_usd_micros = $6,
cash_rounding_unit = $7,
updated_at = now()
WHERE code = $1
RETURNING *;

-- name: GetCurrency :one
SELECT * FROM currencies
WHERE code = $1 LIMIT 1;

-- name: GetDefaultCurrency :one
SELECT * FROM currencies
WHERE is_default = true;

-- name: ListCurrencies :many
-- page_size = 0 returns every currency (the CurrencySelect picker needs the
-- whole list); any positive value pages. Order is always by code, so the
-- picker and the paged table agree.
SELECT * FROM currencies
ORDER BY code
LIMIT NULLIF(sqlc.arg(page_size)::int, 0)
OFFSET sqlc.arg(page_offset);

-- name: CountCurrencies :one
SELECT COUNT(*) FROM currencies;

-- name: ListCurrenciesUpdatedSince :many
SELECT * FROM currencies
WHERE updated_at > $1
ORDER BY updated_at;
