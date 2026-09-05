-- name: CreateCouponCategory :one
INSERT INTO coupon_categories (
  name,
  is_active,
  icon,
  color,
  scope
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetCouponCategory :one
SELECT * FROM coupon_categories
WHERE id = $1 LIMIT 1;

-- name: ListCouponCategories :many
SELECT * FROM coupon_categories
WHERE (sqlc.narg(scope)::text IS NULL OR scope = sqlc.narg(scope))
ORDER BY id;

-- name: UpdateCouponCategory :one
UPDATE coupon_categories
SET name = $2,
    is_active = $3,
    icon = $4,
    color = $5,
    scope = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCouponCategory :exec
DELETE FROM coupon_categories
WHERE id = $1;
