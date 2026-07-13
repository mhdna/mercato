create table if not exists "barcodes" (
  barcode bigint not null unique,
  product_id bigint not null references products(id),
  -- color_id bigint references colors(id),
  -- size_id bigint references sizes(id),
  version int NOT NULL DEFAULT 1,
  created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW()
);

-- unique barcode per product & color & size
CREATE UNIQUE INDEX barcodes_unique_variant
  ON barcodes (product_id); --, COALESCE(color_id, 0), COALESCE(size_id, 0));

CREATE SEQUENCE barcode_item_seq;

SELECT setval('barcode_item_seq',
  COALESCE(
    (SELECT max((barcode / 10) % 1000000000) FROM barcodes),
    0
  ) + 1,
  false
);
