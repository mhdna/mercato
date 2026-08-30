import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

export function useBranchInvoices () {
  async function listBranchInvoiceItems (invoiceId) {
    const data = await requestJSON(`${API_BASE}/branch_invoices/${invoiceId}/items`)
    return data?.items ?? []
  }

  // Most recently received branch invoices -- ordered by id DESC server-side,
  // which tracks received_at since ids are assigned in receipt order. Used by
  // SyncCard to show real "last synced" data instead of a placeholder.
  async function listRecentBranchInvoices (pageSize = 5) {
    const data = await requestJSON(`${API_BASE}/branch_invoices?page_size=${pageSize}&page_id=0`)
    return data?.branch_invoices ?? []
  }

  // Enqueues a remote "remote_return_invoice" command (see
  // api/branch_command.go) asking the branch to process a return itself,
  // using its own real local IDs -- kashi never fabricates a branch invoice
  // row directly. Returns as soon as kashi has queued it; the branch
  // executes asynchronously and the resulting real return invoice arrives
  // later through the normal /branch/return_invoices sync path, same as
  // any other branch-originated return.
  //
  // `items` is optional -- a list of { branch_product_id, quantity } for
  // the specific lines/quantities to return (branch_product_id comes
  // straight from listBranchInvoiceItems, which is the branch's own local
  // product id -- kashi never reinterprets it). Omitted/empty returns
  // every line on the invoice at full quantity, same as before item-level
  // selection existed.
  async function requestRemoteReturn (branchId, invoice, reason, items = []) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/commands`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type: 'remote_return_invoice',
        payload: {
          branch_invoice_id: invoice.id,
          client_ref: invoice.client_ref,
          reason: reason || '',
          items,
        },
      }),
    })
    return data?.command ?? null
  }

  return { listBranchInvoiceItems, listRecentBranchInvoices, requestRemoteReturn }
}
