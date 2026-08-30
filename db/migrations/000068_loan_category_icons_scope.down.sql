alter table loan_categories drop constraint if exists loan_categories_scope_check;
alter table loan_categories drop column if exists updated_at;
alter table loan_categories drop column if exists scope;
alter table loan_categories drop column if exists color;
alter table loan_categories drop column if exists icon;
