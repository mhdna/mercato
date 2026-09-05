alter table assets rename column category_id to type_id;

alter table asset_categories drop constraint if exists asset_categories_name_key;
alter table asset_categories drop column if exists created_at;
alter table asset_categories drop column if exists color;
alter table asset_categories drop column if exists icon;
alter table asset_categories drop column if exists is_active;
alter table asset_categories rename column name to type;
alter table asset_categories rename to assets_types;
