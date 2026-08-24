-- name: CreateCashbox :one
INSERT INTO cashboxes (
  code,
  name,
  is_active
) 
VALUES ( $1, $2, $3 )
RETURNING *;

-- name: GetCashbox :one
SELECT * FROM cashboxes
WHERE id = $1 LIMIT 1;

-- name: ListCashboxes :many
SELECT * FROM cashboxes
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateCashbox :one
UPDATE cashboxes
SET code = $2,
name = $3,
is_active = $4
WHERE id = $1
RETURNING *;

-- name: CreateCashboxAccount :one
INSERT INTO cashbox_accounts (
  name,
  currency_code,
  sort_order,
  color
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCashboxAccount :one
SELECT * FROM cashbox_accounts
WHERE id = $1
LIMIT 1;

-- name: UpdateCashboxAccount :one
UPDATE cashbox_accounts
SET name = $2,
currency_code = $3,
sort_order = $4,
color = $5,
updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListCashboxAccountsUpdatedSince :many
SELECT * FROM cashbox_accounts
WHERE updated_at > $1
ORDER BY updated_at;

-- name: GetCashboxAccountBalance :one
SELECT * FROM shifts_accounts_balances
WHERE account_id = $1
AND shift_id = $2;

-- name: AddCashboxAccountBalance :one
INSERT INTO shifts_accounts_balances (account_id, shift_id, balance)
VALUES ($1, $2, $3)
ON CONFLICT (account_id, shift_id)
DO UPDATE SET balance = shifts_accounts_balances.balance + $3
RETURNING *;

-- name: ListCashboxAccounts :many
SELECT * FROM cashbox_accounts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CreateSalesperson :one
INSERT INTO salespersons (
  name,
  cashbox_id
)
VALUES ( $1, $2 )
RETURNING *;

-- name: GetSalesperson :one
SELECT * FROM salespersons
WHERE id = $1 LIMIT 1;

-- name: ListSalespersons :many
SELECT * FROM salespersons
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListSalespersonsByCashbox :many
SELECT * FROM salespersons
WHERE cashbox_id = $1
ORDER BY id;

-- name: ListAllCashboxAccounts :many
SELECT * FROM cashbox_accounts
ORDER BY id;

-- name: UpdateSalesperson :one
UPDATE salespersons
SET name = $2,
cashbox_id = $3
WHERE id = $1
RETURNING *;

-- name: DeleteSalesperson :exec
DELETE FROM salespersons
WHERE id = $1;
