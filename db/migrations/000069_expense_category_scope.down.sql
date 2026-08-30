alter table expense_categories drop constraint if exists expense_categories_scope_check;
alter table expense_categories drop column if exists updated_at;
alter table expense_categories drop column if exists scope;
