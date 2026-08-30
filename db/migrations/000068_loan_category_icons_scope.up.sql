-- Bring loan_categories up to the same shape expense_categories already
-- has (000058): an icon + accent colour so the admin UI can render them
-- as an icon list. Plus `scope` -- 'central' categories are managed only
-- in kashi; 'branch' categories are the lender-source names branches
-- report through kashi-pos and are synced back down to it (icon + colour
-- included). updated_at backs the sync catch-up cursor -- explicit on
-- every write, no trigger, matching currencies (000026).
alter table loan_categories add column if not exists icon text not null default 'mdi-tag-outline';
alter table loan_categories add column if not exists color text not null default 'blue-grey';
alter table loan_categories add column if not exists scope text not null default 'central';
alter table loan_categories add column if not exists updated_at timestamp(0) with time zone not null default now();

-- Existing rows keep the 'central' default; getOrCreateLoanCategoryID
-- (branch_loan.go) starts stamping 'branch' on categories reported by a
-- branch from here on, so no backfill is needed.
alter table loan_categories drop constraint if exists loan_categories_scope_check;
alter table loan_categories add constraint loan_categories_scope_check check (scope in ('central', 'branch'));
