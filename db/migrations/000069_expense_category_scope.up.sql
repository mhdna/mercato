-- expense_categories already has icon/color (000058); this adds `scope`
-- (same 'central' vs 'branch' split as loan_categories in 000068) and the
-- updated_at column the branch sync catch-up cursor needs. Existing rows
-- keep the 'central' default; getOrCreateExpenseCategoryID (branch_expense.go)
-- stamps 'branch' on categories a branch reports from here on.
alter table expense_categories add column if not exists scope text not null default 'central';
alter table expense_categories add column if not exists updated_at timestamp(0) with time zone not null default now();

alter table expense_categories drop constraint if exists expense_categories_scope_check;
alter table expense_categories add constraint expense_categories_scope_check check (scope in ('central', 'branch'));
