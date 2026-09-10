-- Reconcile queries: given the refs a branch believes it has pushed, return
-- the subset kashi actually holds. The branch diffs its own list against
-- this to decide what still needs (re)queuing -- the basis of "sync
-- everything to a (possibly new) central server". Each is a single indexed
-- `= ANY` lookup; the handler caps the input array size.

-- name: BranchInvoiceRefsPresent :many
SELECT client_ref FROM branch_invoices
WHERE branch_id = $1 AND client_ref = ANY($2::text[]);

-- name: BranchExpenseRefsPresent :many
SELECT client_ref FROM branch_expenses
WHERE branch_id = $1 AND client_ref = ANY($2::text[]);

-- name: BranchLoanRefsPresent :many
SELECT client_ref FROM loans
WHERE branch_id = $1 AND client_ref = ANY($2::text[]);

-- name: BranchClientLinksPresent :many
SELECT branch_client_id FROM client_links
WHERE branch_id = $1 AND branch_client_id = ANY($2::bigint[]);

-- name: ProductBarcodesPresent :many
SELECT barcode FROM product_variants
WHERE barcode = ANY($1::text[]);
