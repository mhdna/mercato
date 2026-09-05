-- name: CreateCoupon :one
INSERT INTO coupons (
  code,
  status,
  discount_type,
  reason,
  client_id,
  valid_until,
  category_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetCoupon :one
SELECT * FROM coupons
WHERE code = $1 LIMIT 1;

-- name: UpdateCoupon :one
UPDATE coupons
SET status = $2,
    discount_type = $3,
    reason = $4,
    client_id = $5,
    valid_until = $6,
    category_id = $7
WHERE code = $1
RETURNING *;

-- name: ListCoupons :many
SELECT * FROM coupons
WHERE (
    BTRIM(sqlc.arg(search)::text) = ''
    OR code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
    OR reason ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  )
  AND (sqlc.narg(status)::coupon_status IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(category_id)::bigint IS NULL OR category_id = sqlc.narg(category_id))
ORDER BY code
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: DeactivateCoupon :exec
UPDATE coupons
  SET status = 'inactive'
WHERE code = $1;

-- name: DeleteCoupons :execrows
DELETE FROM coupons
WHERE code = ANY(sqlc.arg(codes)::text[]);

-- name: CountCoupons :one
SELECT COUNT(*) FROM coupons
WHERE (
    BTRIM(sqlc.arg(search)::text) = ''
    OR code ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
    OR reason ILIKE '%' || BTRIM(sqlc.arg(search)::text) || '%'
  )
  AND (sqlc.narg(status)::coupon_status IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(category_id)::bigint IS NULL OR category_id = sqlc.narg(category_id));
