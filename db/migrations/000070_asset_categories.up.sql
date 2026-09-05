-- Give assets a proper category taxonomy, mirroring loan_categories /
-- expense_categories (000058/000068): an icon + accent colour so the admin
-- UI can render them as an icon list, plus is_active and created_at. The
-- old assets_types table only held a bare `type` label with insert/delete
-- and no way to edit it.
alter table assets_types rename to asset_categories;
alter table asset_categories rename column type to name;
alter table asset_categories add column if not exists is_active boolean not null default true;
alter table asset_categories add column if not exists icon text not null default 'mdi-tag-outline';
alter table asset_categories add column if not exists color text not null default 'blue-grey';
alter table asset_categories add column if not exists created_at timestamp(0) with time zone not null default now();
alter table asset_categories add constraint asset_categories_name_key unique (name);

-- assets.type_id -> category_id, to match loans.category_id / the UI's
-- categoryFor()/CategoryChip conventions. The FK follows the rename.
alter table assets rename column type_id to category_id;
