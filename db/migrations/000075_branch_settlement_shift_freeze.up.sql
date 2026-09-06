-- Once a branch shift is closed in kashi-pos its sales are final: a later
-- "change settlement" edit for one of those older invoices must not rewrite
-- kashi's record of who-paid-with-what. kashi deliberately does not care
-- about post-close edits to a closed shift's invoices.
--
-- Such a settlement is still stored (so the branch's sync outbox gets its
-- 200 and stops retrying) but is flagged and never replayed onto
-- branch_invoice_payments. Settlements for the still-open shift keep
-- applying, and keep surfacing as a change, exactly as before.
--
--   applied              -> replayed onto the sale's payment split
--   pending              -> the sale isn't on the cloud yet; recorded, not replayed
--   ignored_shift_closed -> the sale's shift was already closed; recorded, not replayed
--
-- Existing rows default to 'applied': the historical settlements that ran
-- before this column existed had already been replayed.
alter table branch_invoice_settlements
  add column if not exists status text not null default 'applied'
    constraint branch_invoice_settlements_status_check
      check (status in ('applied', 'pending', 'ignored_shift_closed'));
