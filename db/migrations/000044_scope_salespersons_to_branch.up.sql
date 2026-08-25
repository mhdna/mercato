-- Salespersons become a branch-scoped, cloud-managed roster (synced down to
-- kashi-pos read-only, same as branch_targets/currencies/cashbox_accounts).
-- cashbox_id is left untouched -- it's unrelated legacy scoping still read
-- by api/pos_settings.go, and the new branch-scoped path doesn't use it.
alter table salespersons
    add column branch_id bigint references branches(id) on delete cascade,
    add column is_active boolean not null default true,
    add column updated_at timestamp(0) with time zone not null default now();

create index on salespersons (branch_id);
