-- Branch-reported expenses, following the same append-only, opaque-branch-ID
-- pattern as branch_invoices (see 000025_create_branch_invoices.up.sql):
-- branch_cashbox_account_id/branch_shift_id are the branch's own local
-- SQLite ids, meaningless to kashi's schema, stored as plain traceability
-- data rather than foreign keys until branch-scoped identity sync exists.
create table if not exists branch_expenses (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    description text not null,
    category text not null default '',
    amount bigint not null,
    currency_code text not null references currencies(code),

    branch_cashbox_account_id bigint not null,
    branch_shift_id bigint,

    occurred_at timestamp(0) with time zone not null,
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);
