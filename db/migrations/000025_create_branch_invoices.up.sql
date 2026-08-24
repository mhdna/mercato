-- Branch-submitted invoices are stored separately from kashi's own
-- invoices table, not merged into it. Two problems make that necessary:
--
-- 1. invoices.cashbox_id/shift_id/client_id/inventory_id are NOT NULL
--    foreign keys into kashi's own tables, but a branch's cashbox/shift/
--    client/inventory/product IDs are local SQLite autoincrement numbers
--    private to that one install — they don't resolve to any kashi row
--    (and won't, until product/client/branch-scoped-cashbox sync exists).
-- 2. kashi's own SalesInvoiceTx/ReturnInvoiceTx independently recompute
--    loyalty points and cashbox-account balances from grand_total using
--    kashi's own formula. That formula already disagrees with kashi-pos's
--    (grandTotal vs grandTotal/100 for loyalty points) — recomputing here
--    would silently overwrite what the branch's till already decided and
--    already showed the cashier, which must remain the source of truth.
--
-- So this table is a faithful, append-only record of what a branch
-- reports, not an attempt to integrate into kashi's normalized invoice/
-- entry/balance bookkeeping. That integration is later work, once branch
-- identity for clients/cashboxes/products actually exists centrally.
create table if not exists branch_invoices (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    kind text not null check (kind in ('sales', 'return')),
    branch_invoice_code text not null,

    -- Branch-local identifiers, stored as opaque data for traceability —
    -- explicitly NOT foreign keys, since they don't resolve to any kashi
    -- row today. kashi-pos has no separate "cashbox register" entity of
    -- its own — invoices.cashbox_id there is actually a cashbox_account
    -- (payment method) id, so there's deliberately no parallel
    -- branch_cashbox_id column here either.
    branch_cashbox_account_id bigint not null,
    branch_shift_id           bigint not null,
    branch_inventory_id       bigint not null,
    branch_client_id          bigint,

    -- For a return: the client_ref of the original sale within the same
    -- branch. Not a real FK for the same reason as above (the referenced
    -- row is in this same table, but scoped by branch_id, not a plain id).
    related_client_ref text,

    discount smallint not null check (discount >= 0 and discount <= 100),
    subtotal bigint not null,
    discounted_total bigint not null,
    grand_total bigint not null,

    -- Reported by the branch, not recomputed here — see note above.
    loyalty_points_delta bigint not null default 0,

    -- When the sale actually happened at the branch, which may be long
    -- before kashi received it (offline branch catching up).
    occurred_at timestamp(0) with time zone not null,
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);

create table if not exists branch_invoice_items (
    id bigserial primary key,
    branch_invoice_id bigint not null references branch_invoices(id) on delete cascade,
    branch_product_id bigint not null,
    unit_price bigint not null,
    line_total bigint not null,
    discount smallint not null check (discount >= 0 and discount <= 100),
    quantity bigint not null
);
