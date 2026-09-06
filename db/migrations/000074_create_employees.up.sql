-- Central HR roster plus a single current payroll snapshot per employee.
-- Global, not branch-scoped: an employee may optionally be attributed to a
-- branch (branch_id, nullable) for reporting, but the admin list is
-- company-wide. All money is USD integer cents, matching branch_shifts /
-- branch_invoices -- no `numeric` columns anywhere in this schema.
--
-- The attendance/late deduction (حسم) is modelled as a rate the admin
-- quotes per unit (late_deduction_rate_cents, in late_deduction_unit of
-- 'hour' or 'day') times a quantity of late units the admin enters
-- (late_units_centi -- hundredths of a unit, so 2.5 hours is stored 250,
-- since the schema has no fractional numeric type). The deduction
-- (rate * units / 100) and the net salary (base + commission - deduction)
-- are pure arithmetic recomputed by the API on read, never stored.
create table if not exists employees (
    id bigserial primary key,
    branch_id bigint references branches(id) on delete set null,
    name text not null,
    role text not null default '',
    status text not null default 'active'
        constraint employees_status_check check (status in ('active', 'inactive')),
    hired_on date,

    base_salary_cents bigint not null default 0
        constraint employees_base_salary_check check (base_salary_cents >= 0),
    commission_cents bigint not null default 0
        constraint employees_commission_check check (commission_cents >= 0),
    late_deduction_rate_cents bigint not null default 0
        constraint employees_late_rate_check check (late_deduction_rate_cents >= 0),
    late_deduction_unit text not null default 'hour'
        constraint employees_late_unit_check check (late_deduction_unit in ('hour', 'day')),
    late_units_centi bigint not null default 0
        constraint employees_late_units_check check (late_units_centi >= 0),

    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now()
);

create index on employees (branch_id);
create index on employees (name);
