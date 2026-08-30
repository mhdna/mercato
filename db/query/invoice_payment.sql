-- name: CreateInvoicePayment :one
INSERT INTO invoice_payments (invoice_id, cashbox_account_id, amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListInvoicePaymentsByInvoice :many
SELECT invoice_payments.*, cashbox_accounts.name AS account_name
FROM invoice_payments
JOIN cashbox_accounts ON cashbox_accounts.id = invoice_payments.cashbox_account_id
WHERE invoice_payments.invoice_id = $1
ORDER BY invoice_payments.id;
