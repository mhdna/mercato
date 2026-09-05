-- Admin-managed taxonomy for coupons, mirroring loan_categories (000037 +
-- 000068): an icon + accent colour for the admin icon list, plus a `scope`
-- column kept only for parity with the shared CategoryDialog /
-- CategorySidebar UI -- coupons are central-only, so rows are always
-- 'central'. updated_at is stamped explicitly on every write (no trigger),
-- matching loan_categories.
create table if not exists coupon_categories (
    id bigserial primary key,
    name text not null unique,
    is_active boolean not null default true,
    icon text not null default 'mdi-tag-outline',
    color text not null default 'blue-grey',
    scope text not null default 'central',
    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now(),
    constraint coupon_categories_scope_check check (scope in ('central', 'branch'))
);

-- Nullable: existing coupons and the "uncategorised" case are both valid.
-- No ON DELETE CASCADE, so a category still referenced by a coupon can't be
-- deleted -- the handler surfaces that FK violation as a 409.
alter table coupons add column if not exists category_id bigint references coupon_categories(id);
