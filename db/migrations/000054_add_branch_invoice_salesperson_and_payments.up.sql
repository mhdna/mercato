ALTER TABLE branch_invoices
  ADD COLUMN salesperson_name TEXT NOT NULL DEFAULT '';

-- One row per payment account a branch invoice was settled through, as
-- reported by kashi-pos (already computed locally there from its own
-- invoice_payments table -- see sync_outbox.go's buildBranchInvoicePayments
-- in the kashi-pos repo). account_name is a display name, not an id:
-- cashbox accounts aren't a synced/shared entity across the branch
-- boundary, same reasoning as branch_product_id staying opaque.
CREATE TABLE branch_invoice_payments (
  id BIGSERIAL PRIMARY KEY,
  branch_invoice_id BIGINT NOT NULL REFERENCES branch_invoices (id) ON DELETE CASCADE,
  account_name TEXT NOT NULL,
  amount BIGINT NOT NULL
);

CREATE INDEX ON branch_invoice_payments (branch_invoice_id);
