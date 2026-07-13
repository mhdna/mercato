create table if not exists price_lists (
    id bigserial primary key,
    name text not null,
    is_active boolean not null default true,
    is_default boolean not null default false,
    valid_from timestamp(0) with time zone not null default now(),
    valid_to timestamp(0) with time zone not null default now(),
    created_at timestamp(0) with time zone not null default now()
);

create table if not exists price_list_items (
    price_list_id bigint not null references price_lists(id) on delete cascade,
    product_id bigint not null references products(id) on delete cascade,
    price bigint not null,
    primary key (price_list_id, product_id)
);

create table if not exists discount_lists (
    id bigserial primary key,
    name text not null,
    is_active boolean not null default true,
    is_default boolean not null default false,
    valid_from timestamp(0) with time zone not null default now(),
    valid_to timestamp(0) with time zone not null default now(),
    created_at timestamp(0) with time zone not null default now()
);

create table if not exists discount_list_items (
    discount_list_id bigint not null references discount_lists(id) on delete cascade,
    product_id bigint not null references products(id) on delete cascade,
    discount smallint not null check (discount >= 0 and discount <= 100),
    primary key (discount_list_id, product_id)
);

-- ALTER TABLE invoices ADD COLUMN price_list_id bigint REFERENCES price_lists(id);
