create table if not exists branches (
    id bigserial primary key,
    name text not null,
    code text not null unique,
    api_key_hash text not null,
    is_active boolean not null default true,
    last_seen_at timestamp(0) with time zone,
    created_at timestamp(0) with time zone not null default now()
);
