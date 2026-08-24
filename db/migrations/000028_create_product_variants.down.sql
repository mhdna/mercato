create table if not exists "barcodes" (
  barcode bigint not null unique,
  product_id bigint not null references products(id),
  version int NOT NULL DEFAULT 1,
  created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX barcodes_unique_variant ON barcodes (product_id);

insert into barcodes (barcode, product_id, created_at)
select barcode::bigint, product_id, created_at from product_variants
where barcode ~ '^[0-9]+$';

create table if not exists "products_colors" (
  product_id bigint not null references products(id) on delete cascade,
  color_id bigint not null references colors(id) on delete cascade,
  primary key (product_id, color_id)
);

create table if not exists products_sizes (
  product_id bigint not null references products(id) on delete cascade,
  size_id bigint not null references sizes(id) on delete cascade,
  primary key (product_id, size_id)
);

drop table if exists product_variants;
