-- Repayments against a loan (سحوبات) -- recorded centrally only, against
-- any loan regardless of its origin. No branch reporting path for these,
-- unlike loans themselves.
create table if not exists loan_payments (
    id bigserial primary key,
    loan_id bigint not null references loans(id),
    amount bigint not null,
    currency_code text not null references currencies(code),
    note text not null default '',
    paid_at timestamp(0) with time zone not null default now()
);

create index on loan_payments (loan_id);
