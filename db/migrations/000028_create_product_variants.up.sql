-- Replaces the never-finished variant model (barcodes: one per product,
-- with color_id/size_id commented out since the migration was written;
-- products_colors/products_sizes: many-to-many "comes in these
-- colors/sizes" joins, not a real per-combination row) with a real
-- variant: one row per sellable SKU, with its own barcode and price.
--
-- product_id + color_id + size_id is unique so kashi-pos's own
-- (code, color, size) uniqueness invariant (000027_product_variants.up.sql
-- there) has a direct counterpart here.
create table if not exists product_variants (
    id bigserial primary key,
    product_id bigint not null references products(id) on delete cascade,
    color_id bigint references colors(id),
    size_id bigint references sizes(id),
    barcode text not null unique,
    price bigint,
    is_active boolean not null default true,
    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now(),
    unique (product_id, color_id, size_id)
);

-- products_colors/products_sizes are empty in every environment this has
-- run in (the feature was never actually used to assign a color/size to a
-- product) — there is no many-to-many data to reconcile into per-variant
-- rows, so this is a straight carry-over of the one thing that does have
-- real rows: each product's existing barcode.
insert into product_variants (product_id, barcode, created_at)
select product_id, barcode::text, created_at from barcodes;

drop table if exists barcodes;
drop table if exists products_colors;
drop table if exists products_sizes;
