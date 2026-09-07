-- Branch-reported visitor-counter presses. Like branch_shifts (see 000025 /
-- the branch_shifts migration), this is an append-only, faithful record of
-- what a branch's till reported -- one row per press of kashi-pos's
-- nav-drawer people counter -- not an integration into any central entity.
-- kashi has no visitor/footfall concept of its own; the admin UI just rolls
-- these up per branch per day.
--
-- branch_id is a real foreign key (the branch is a central entity);
-- client_ref is the kashi-pos-minted uuid, unique per (branch_id,
-- client_ref) so a retried sync_outbox entry is a no-op. direction is
-- 'in' (counter ▲) or 'out' (counter ▼). occurred_at is the till's local
-- press time; the branch also sends the "YYYY-MM-DD" day it bucketed the
-- press under so the rollup matches the till's own without kashi
-- re-deriving a calendar day from a timezone-bearing timestamp.
create table if not exists branch_visitor_events (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    day text not null,
    direction text not null check (direction in ('in', 'out')),

    occurred_at timestamp(0) with time zone not null,
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);

create index on branch_visitor_events (branch_id, day);
