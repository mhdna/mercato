-- A branch's POS user roster, admin-managed centrally and synced down to
-- kashi-pos read-only -- same shape as the branch-scoped salespersons
-- roster (000044). The PIN is deliberately NOT stored here: kashi hashes
-- with bcrypt, kashi-pos with sha256, and a 4-6 digit PIN has little value
-- at rest. PIN set/reset goes through a `set_branch_user_pin` branch
-- command carrying the plaintext, which the till hashes locally.
create table if not exists branch_users (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    username text not null,
    role text not null default 'cashier' check (role in ('admin', 'cashier')),
    akuvox_user_id text not null default '',
    is_active boolean not null default true,
    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now(),
    unique (branch_id, username)
);

create index on branch_users (branch_id);
