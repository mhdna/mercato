-- A branch-reported "change settlement" edit: the new per-account payment
-- split for a sale that was already synced to kashi. client_ref here is the
-- edit's own settlement_ref -- each edit is a distinct event, deduplicated
-- on (branch_id, client_ref), never collapsed with the sale's prior
-- settlement. sale_client_ref is the branch_invoices.client_ref of the sale
-- being re-settled; createBranchInvoiceSettlement replays this split onto
-- that invoice's branch_invoice_payments rows so the admin view of
-- who-paid-with-what stays correct. A settlement whose sale is not on the
-- cloud yet is still stored (this table + its payments), to be reconciled
-- if the sale arrives later.
create table if not exists branch_invoice_settlements (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    sale_client_ref text not null,
    grand_total bigint not null default 0,
    occurred_at timestamp(0) with time zone not null,
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);

create index on branch_invoice_settlements (branch_id, sale_client_ref);

-- One row per payment account the sale is now settled through, same shape
-- and "account_name is display text, not an id" reasoning as
-- branch_invoice_payments (000054).
create table if not exists branch_invoice_settlement_payments (
    id bigserial primary key,
    branch_invoice_settlement_id bigint not null references branch_invoice_settlements(id) on delete cascade,
    account_name text not null,
    amount bigint not null
);

create index on branch_invoice_settlement_payments (branch_invoice_settlement_id);
