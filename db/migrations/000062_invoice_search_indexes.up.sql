CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_invoices_invoice_code_trgm
ON invoices USING gin (invoice_code gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_branch_invoices_code_trgm
ON branch_invoices USING gin (branch_invoice_code gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_branch_invoices_salesperson_trgm
ON branch_invoices USING gin (salesperson_name gin_trgm_ops);
