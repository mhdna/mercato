-- name: GetNextBarcodeItemValue :one
SELECT nextval('barcode_item_seq')::bigint;
