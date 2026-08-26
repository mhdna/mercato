alter table branch_targets drop constraint branch_targets_series_id_date_from_key;
drop index if exists idx_branch_invoices_occurred_at;
drop index if exists idx_cashbox_accounts_updated_at;
drop index if exists idx_currencies_updated_at;
drop index if exists idx_clients_updated_at;
drop index if exists idx_product_variants_updated_at;
