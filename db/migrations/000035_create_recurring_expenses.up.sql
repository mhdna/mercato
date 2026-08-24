-- A recurring expense is a template kashi fires automatically on a
-- day/month interval (e.g. electricity every 5 months), creating a real
-- row in `expenses` each time it comes due -- see the recurring-expense
-- scheduler in main.go. `active` lets an admin pause firing without
-- deleting the template (and losing next_due_at's schedule position).
create table if not exists recurring_expenses (
    id bigserial primary key,
    description text not null,
    category text not null default '',
    amount bigint not null,
    currency_code text not null references currencies(code),

    interval_unit text not null check (interval_unit in ('day', 'month')),
    interval_count int not null check (interval_count > 0),

    next_due_at timestamp(0) with time zone not null,
    active boolean not null default true,

    created_at timestamp(0) with time zone not null default now()
);
