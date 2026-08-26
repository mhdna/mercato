alter table branch_invoices drop constraint branch_invoices_kind_check;
alter table branch_invoices add constraint branch_invoices_kind_check
    check (kind in ('sales', 'return'));
