-- branchSyncChanges (api/branch_catchup.go) runs a "WHERE updated_at > $1"
-- query against each of these tables on every branch's poll tick -- with no
-- index, that's a recurring full scan multiplied by however many branches
-- are polling. All four already have this shape of query in production
-- (ListProductVariantsForSync, ListClientsUpdatedSince,
-- ListCurrenciesUpdatedSince, ListCashboxAccountsUpdatedSince).
create index if not exists idx_product_variants_updated_at on product_variants (updated_at);
create index if not exists idx_clients_updated_at on clients (updated_at);
create index if not exists idx_currencies_updated_at on currencies (updated_at);
create index if not exists idx_cashbox_accounts_updated_at on cashbox_accounts (updated_at);

-- branch_invoices is append-only and only grows; ListDailyIncome filters by
-- occurred_at (see the companion query rewrite making that filter sargable)
-- and the admin branch-invoices page filters by branch_id, neither of which
-- had a standalone index (branch_id was only indexed as the leading column
-- of the (branch_id, client_ref) uniqueness constraint, which doesn't help
-- a plain occurred_at range scan).
create index if not exists idx_branch_invoices_occurred_at on branch_invoices (occurred_at);

-- fireDueBranchTargetSeries (branch_target_series_scheduler.go) does a
-- check-then-create per due series on every scheduler tick with no DB-level
-- backstop -- a genuine race (e.g. ever running two scheduler instances)
-- could insert duplicate target rows for the same period. Postgres unique
-- indexes treat NULLs as distinct, so this doesn't block the many one-off,
-- non-recurring targets that have series_id = NULL.
alter table branch_targets add constraint branch_targets_series_id_date_from_key unique (series_id, date_from);
