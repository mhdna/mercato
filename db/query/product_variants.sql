-- name: CreateProductColor :one
INSERT INTO products_colors (
  product_id,
  color_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: CreateProductSize :one
INSERT INTO products_sizes (
  product_id,
  size_id
) VALUES (
    $1, $2
) RETURNING *;
