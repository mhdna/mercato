-- name: CreateCoupon :one
INSERT INTO coupons (
  code,
  status,
  discount_type,
  reason,
  client_id,
  valid_until
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetCoupon :one
SELECT * FROM coupons
WHERE code = $1 LIMIT 1;

-- name: ListCoupons :many
SELECT * FROM coupons
WHERE BTRIM(sqlc.arg(search)::text) = ''
   OR code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
   OR reason ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
ORDER BY code
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: DeactivateCoupon :exec
UPDATE coupons
  SET status = 'inactive'
WHERE code = $1;

-- name: CountCoupons :one
SELECT COUNT(*) FROM coupons
WHERE BTRIM(sqlc.arg(search)::text) = ''
   OR code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
   OR reason ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%';
