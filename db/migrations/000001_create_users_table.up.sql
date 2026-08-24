CREATE EXTENSION IF NOT EXISTS citext;

create table if not exists users (
    id bigserial primary key,
    name text not null,
    email citext unique not null,
    -- password_hash bytea not null,
    password_hash varchar not null,
    password_changed_at timestamp(0) with time zone not null default '0001-01-01 00:00:00+00',
    activated bool not null,
    version int not null default 1,
    created_at timestamp(0) with time zone not null default now()
);
