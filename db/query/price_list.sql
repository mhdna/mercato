-- name: CreatePriceList :one
INSERT INTO price_lists (
  name,
  is_active,
  is_default,
  valid_from,
  valid_to
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPriceList :one
SELECT * FROM price_lists
WHERE id = $1 LIMIT 1;

-- name: ListPriceLists :many
SELECT * FROM price_lists
WHERE BTRIM(sqlc.arg(search)::text) = ''
   OR name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
ORDER BY id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountPriceListsFiltered :one
SELECT COUNT(*) FROM price_lists
WHERE BTRIM(sqlc.arg(search)::text) = ''
   OR name ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%';

-- name: UpdatePriceList :exec
UPDATE price_lists
SET name = $2, is_active = $3, is_default = $4, valid_from = $5, valid_to = $6
WHERE id = $1;

-- name: UnsetDefaultPriceList :exec
UPDATE price_lists
SET is_default = false
WHERE is_default = true AND id != $1;

-- name: DeletePriceList :exec
DELETE FROM price_lists
WHERE id = $1;
