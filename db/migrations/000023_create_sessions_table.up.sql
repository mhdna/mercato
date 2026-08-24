create table if not exists sessions (
    id uuid primary key,
    username text not null,
    refresh_token text not null,
    user_agent text not null,
    client_ip text not null,
    is_blocked boolean not null default false,
    expires_at timestamp(0) with time zone not null,
    created_at timestamp(0) with time zone not null default now()
);
