-- Admin-managed categories for expenses, full CRUD -- replaces expenses'
-- (and branch_expenses'/recurring_expenses') free-text category field with
-- a controlled taxonomy, same shape as loan_categories. Deliberately
-- unseeded: 000041 backfills it from whatever category text already
-- exists in the data, nothing more.
create table if not exists expense_categories (
    id bigserial primary key,
    name text not null unique,
    is_active boolean not null default true,
    created_at timestamp(0) with time zone not null default now()
);
