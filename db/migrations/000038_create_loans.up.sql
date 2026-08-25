-- A single loans table covers both loans entered directly in kashi (origin
-- = 'central_loan') and loans a branch reports through kashi-pos (origin =
-- 'branch_loan'), distinguished by the indexed `origin` column rather than
-- a separate branch_loans table like branch_expenses -- kashi-pos POSTs
-- branch loans straight into this table via /branch/loans, using the same
-- idempotent (branch_id, client_ref) pattern as branch_expenses.
-- branch_cashbox_account_id/branch_shift_id are the branch's own local
-- SQLite ids, meaningless to kashi's schema, kept as plain traceability
-- data rather than foreign keys, same reasoning as branch_expenses.
create table if not exists loans (
    id bigserial primary key,
    origin text not null check (origin in ('central_loan', 'branch_loan')),
    description text not null,
    category_id bigint not null references loan_categories(id),
    amount bigint not null,
    currency_code text not null references currencies(code),

    branch_id bigint references branches(id),
    client_ref text,
    branch_cashbox_account_id bigint,
    branch_shift_id bigint,

    occurred_at timestamp(0) with time zone not null default now(),
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref),
    check (
        (origin = 'branch_loan' and branch_id is not null and client_ref is not null and branch_cashbox_account_id is not null)
        or
        (origin = 'central_loan' and branch_id is null and client_ref is null)
    )
);

create index on loans (origin);
create index on loans (branch_id);
