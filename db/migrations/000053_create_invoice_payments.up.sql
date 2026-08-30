-- One row per cashbox account an invoice was settled through. Today kashi's
-- own local invoice creation only ever writes a single row (one account
-- picked in the create form), but the shape supports a real multi-account
-- split -- which branch invoices (synced from kashi-pos, where a checkout
-- can genuinely be split across payment methods) do use.
CREATE TABLE invoice_payments (
  id BIGSERIAL PRIMARY KEY,
  invoice_id BIGINT NOT NULL REFERENCES invoices (id) ON DELETE CASCADE,
  cashbox_account_id BIGINT NOT NULL REFERENCES cashbox_accounts (id),
  amount BIGINT NOT NULL
);

CREATE INDEX ON invoice_payments (invoice_id);
