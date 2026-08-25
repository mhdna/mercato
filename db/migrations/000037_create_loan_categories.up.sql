-- Lender-source categories for loans (e.g. "Owner" vs an external lender),
-- admin-managed via full CRUD -- unlike expenses' free-text category field,
-- this is a controlled taxonomy referenced by loans.category_id.
create table if not exists loan_categories (
    id bigserial primary key,
    name text not null unique,
    is_active boolean not null default true,
    created_at timestamp(0) with time zone not null default now()
);

insert into loan_categories (name) values ('Owner'), ('External');
