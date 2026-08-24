-- name: GetProductAttributes :many
SELECT * FROM products_attributes
WHERE product_id = $1;

-- name: CreateProductAttribute :one
INSERT INTO products_attributes (
  product_id,
  attribute_id,
  attribute_value_id
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetProductAttributeValue :one
SELECT * FROM products_attributes
WHERE product_id = $1 AND attribute_id = $2;

-- name: ListProductAttributes :many
SELECT pa.*
FROM products p
INNER JOIN products_attributes pa
ON p.id = pa.product_id
ORDER BY attribute_id
LIMIT $1
OFFSET $2;

-- name: UpdateProductAttribute :exec
UPDATE products_attributes
SET attribute_value_id = $3
WHERE product_id = $1 AND attribute_id = $2;
