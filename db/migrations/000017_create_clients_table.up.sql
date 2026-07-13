-- -- needed for uuid
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

create table if not exists clients (
    id bigserial primary key,
    -- uuid uuid unique not null default gen_random_uuid(),
    name text not null,
    phone text not null unique,
    total_loyalty_points bigint not null default 0,
    valid_loyalty_points bigint not null default 0,
    created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW()
);

create  type discount_type as enum (
    'fixed',
    'percentage'
);

create  type coupon_status as enum (
    'active',
    'inactive'
);


create table if not exists coupons (
    code text primary key,
    status coupon_status not null,
    discount_type discount_type not null,
    reason text not null,
    client_id bigint not null references clients(id),
    valid_until TIMESTAMP WITH TIME ZONE not null,
    created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW()
);
