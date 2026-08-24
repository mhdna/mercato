-- name: CreateBarcode :one
INSERT INTO barcodes (
  barcode,
  product_id
  -- color_id,
  -- size_id
) VALUES (
    $1, $2 --, $3, $4
) RETURNING *;

-- name: GetNextBarcodeItemValue :one
SELECT nextval('barcode_item_seq')::bigint;

-- name: GetBarcode :one
SELECT * FROM barcodes
WHERE product_id = $1
-- can't use = since color & size are nullable
-- AND color_id IS NOT DISTINCT FROM $2
-- AND size_id IS NOT DISTINCT FROM $3
LIMIT 1;

-- name: ListBarcodes :many
SELECT b.*, p.* FROM barcodes as b
INNER JOIN products as p
-- we have color & sizes in both tables, so them only once instead of twice
-- LEFT JOIN colors c ON c.id = b.color_id
-- LEFT JOIN sizes s ON s.id = b.size_id
-- TODO: fix created_at being 2
ON b.product_id = p.id
ORDER BY p.id
LIMIT $1
OFFSET $2;

-- name: UpdateBarcode :exec
UPDATE barcodes
  SET barcode = $2
WHERE product_id = $1;
-- WARN: beware of renaming the $2 above to $4 or sth
-- AND color_id IS NOT DISTINCT FROM $2
-- AND size_id IS NOT DISTINCT FROM $3;

-- name: DeleteBarcode :exec
DELETE FROM barcodes
WHERE product_id = $1;
-- AND color_id IS NOT DISTINCT FROM $2
-- AND size_id IS NOT DISTINCT FROM $3;
