-- We have multiple accounts per cashbox
-- Multiple balances per shift/cashbox
-- And they're stored into accounts_balances_currencies
-- Where each currency has a balance of its own (e.g. 50 USD, 30 EUR)
-- We store total balance which is
-- (50 USD * VALUE_IN_DEFAULT_CURRENCY + 30 EUR * VALUE_IN_DEFAULT_CURRENCY)
-- inside shifts.total_balance (& in the entry as well)

create table if not exists cashboxes (
    id bigserial primary key,
    name text not null unique,
    code text not null unique,
    is_active boolean not null default true,
    created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW()
);

create table if not exists shifts (
    id bigserial primary key,
    is_closed boolean not null default false,
    cashbox_id bigint not null references cashboxes(id),
    created_at timestamp(0) WITH time zone NOT NULL DEFAULT NOW(),
    closed_at timestamp(0) WITH time zone
);

-- Cash, Digital Card, etc.
create table if not exists cashbox_accounts (
    id bigserial primary key,
    name text not null unique
);

create table if not exists shifts_accounts_balances (
    account_id bigint not null references cashbox_accounts(id),
    shift_id bigint not null references shifts(id),
    balance bigint not null,
	primary key (account_id, shift_id)
);

create table if not exists salespersons (
    id bigserial primary key,
    name text not null,
    cashbox_id bigint references cashboxes(id),
    unique (name, cashbox_id)
);
