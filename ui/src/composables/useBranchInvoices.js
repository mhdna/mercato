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

  return { listBranchInvoiceItems, listRecentBranchInvoices }
}
