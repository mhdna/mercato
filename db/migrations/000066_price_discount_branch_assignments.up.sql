-- Which branches a price list / discount list is enabled for. A branch is the
-- primary key here, so it can be assigned to at most one price list and at
-- most one discount list at a time -- reassigning is an explicit action, and
-- two lists can never both be "active" for the same branch.
create table if not exists price_list_branches (
    branch_id bigint primary key references branches(id) on delete cascade,
    price_list_id bigint not null references price_lists(id) on delete cascade,
    created_at timestamptz not null default now()
);
create index if not exists price_list_branches_price_list_id_idx
    on price_list_branches (price_list_id);

create table if not exists discount_list_branches (
    branch_id bigint primary key references branches(id) on delete cascade,
    discount_list_id bigint not null references discount_lists(id) on delete cascade,
    created_at timestamptz not null default now()
);
create index if not exists discount_list_branches_discount_list_id_idx
    on discount_list_branches (discount_list_id);
