-- Branch-reported shift closes. Like branch_invoices / branch_expenses (see
-- 000025), this is an append-only, faithful record of what a branch's till
-- decided at close time, not an integration into kashi's own shift
-- bookkeeping -- kashi has no branch shift entity of its own.
-- branch_shift_id is the branch's own local shifts.id, stored as opaque
-- traceability data, explicitly NOT a foreign key. Each variance_* is
-- derived centrally as counted - expected rather than trusting a
-- branch-reported figure (same "recompute pure arithmetic, never recompute
-- what the till already showed the cashier" reasoning as
-- loyalty_points_delta on branch_invoices).
create table if not exists branch_shifts (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    branch_shift_id bigint not null,

    opened_at timestamp(0) with time zone,
    closed_at timestamp(0) with time zone,
    closing_person_name text not null default '',

    opening_float_usd bigint not null default 0,
    opening_float_lbp bigint not null default 0,

    expected_usd bigint not null default 0,
    expected_lbp bigint not null default 0,
    expected_visa bigint not null default 0,
    expected_whish bigint not null default 0,

    counted_usd bigint not null default 0,
    counted_lbp bigint not null default 0,
    counted_visa bigint not null default 0,
    counted_whish bigint not null default 0,

    variance_usd bigint not null default 0,
    variance_lbp bigint not null default 0,
    variance_visa bigint not null default 0,
    variance_whish bigint not null default 0,

    occurred_at timestamp(0) with time zone not null,
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);

create index on branch_shifts (branch_id, closed_at);
