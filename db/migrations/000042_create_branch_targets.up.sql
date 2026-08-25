-- Admin-set revenue goals for a branch over a date range. Multiple targets
-- can be active/overlapping at once (e.g. a monthly and a quarterly goal
-- shown together as layered progress bars); progress itself is computed on
-- read from branch_invoices rather than stored, so it always reflects live
-- data.
create table if not exists branch_targets (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    date_from date not null,
    date_to date not null,
    target_amount bigint not null,
    -- Hex accent color for this target's progress bar, admin-settable; if
    -- left blank on create, the API assigns one from a fixed palette (see
    -- assignBranchTargetColor in api/branch_target.go) so bars are never
    -- left uncolored/indistinguishable.
    color text not null default '',
    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now()
);

create index on branch_targets (branch_id);
