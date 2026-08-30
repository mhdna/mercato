import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url) {
  const res = await authFetch(url)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

// "Everything about this invoice" views -- see api/invoice_details.go
// (kashi's own local sales/return invoices) and the getBranchInvoiceDetails
// handler in api/branch_invoice.go (invoices synced from a branch).
export function useInvoiceDetails () {
  async function getInvoiceDetails (invoiceId) {
    return requestJSON(`${API_BASE}/invoices/${invoiceId}/details`)
  }

  async function getBranchInvoiceDetails (branchInvoiceId) {
    return requestJSON(`${API_BASE}/branch_invoices/${branchInvoiceId}/details`)
  }

  return { getInvoiceDetails, getBranchInvoiceDetails }
}
