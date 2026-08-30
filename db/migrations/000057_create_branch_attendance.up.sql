-- Branch-reported attendance, append-only.
--
-- branch_attendance_events is the raw IN/OUT punch stream imported from a
-- branch's biometric device. client_ref is the device's own
-- "<device_id>|<event_id>" key -- already unique per branch -- so a replayed
-- sync batch is idempotent row by row.
--
-- branch_attendance_changes covers the manager-approval workflow: a
-- cashier-filed change request (kind='complaint') and a manager's review of
-- one (kind='approval', carrying the approved/rejected status and the
-- resulting before/after times). client_ref is the branch's local
-- "acr-<id>-complaint" / "acr-<id>-review" key.
--
-- event_* / *_time columns are stored as the branch's own text, not parsed
-- timestamps: they're display data echoed back to the admin UI, and the
-- branch's device is the only authority on their format.
create table if not exists branch_attendance_events (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    salesperson_name text not null default '',
    attendance_user_id text not null default '',
    event_date text not null,
    event_time text not null,
    event_at text not null,
    type text not null default '',
    status text not null default '',
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);

create index on branch_attendance_events (branch_id, event_date);

create table if not exists branch_attendance_changes (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    client_ref text not null,
    kind text not null check (kind in ('complaint', 'approval')),
    salesperson_name text not null default '',
    attendance_date text not null,
    original_time text not null default '',
    requested_time text not null default '',
    requested_type text not null default '',
    action text not null default '',
    status text not null default '',
    note text not null default '',
    actor text not null default '',
    occurred_at timestamp(0) with time zone not null,
    received_at timestamp(0) with time zone not null default now(),

    unique (branch_id, client_ref)
);

create index on branch_attendance_changes (branch_id, attendance_date);
