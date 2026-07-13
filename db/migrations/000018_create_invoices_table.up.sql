create table if not exists invoices (
    id bigserial primary key,
    cashbox_id bigint not null references cashboxes(id),
    shift_id bigint not null references shifts(id),
    invoice_code text not null,
    invoice_index bigint not null,
    year int not null,
    client_id bigint not null references clients(id),
    inventory_id bigint not null references inventories(id),
    discount SMALLINT not null CHECK (discount >= 0 AND discount <= 100),
    subtotal bigint not null,
    discounted_total bigint not null,
    grand_total bigint not null,
    created_at timestamp(0) with time zone not null default now()
);

create table if not exists invoice_products (
    invoice_id bigint not null references invoices(id) on delete cascade,
    product_id bigint not null references products(id),
    unit_price bigint not null,
    line_total bigint not null,
    discount SMALLINT not null CHECK (discount >= 0 AND discount <= 100),
    quantity bigint not null,
    primary key (invoice_id, product_id)
);

CREATE TYPE index_type AS ENUM (
     'sales',
     'return'
);

create table if not exists invoice_indexes (
    year int not null,
    cashbox_id bigint not null references cashboxes(id),
    last_index bigint not null,
    type index_type not null,
    primary key (cashbox_id, year)
);

create table if not exists sales_invoices (
    invoice_id bigint primary key references invoices(id) on delete cascade
);

create table if not exists return_invoices (
    invoice_id bigint primary key references invoices(id) on delete cascade,
    sales_invoice_id bigint not null references sales_invoices(invoice_id)
);
