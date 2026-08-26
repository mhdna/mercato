-- kashi-pos already supports exchange invoices (a net signed cash
-- difference, positive/negative/zero -- see tx_exchange.go's
-- netDifference there, and the ListDailyIncome comment in
-- db/query/branch_invoice.sql), but branch_invoices had no valid kind to
-- report one under, so exchanges were being squeezed into 'return' and
-- getting force-negated as if they were a refund.
alter table branch_invoices drop constraint branch_invoices_kind_check;
alter table branch_invoices add constraint branch_invoices_kind_check
    check (kind in ('sales', 'return', 'exchange'));
