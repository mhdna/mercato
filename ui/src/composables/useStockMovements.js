import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.status === 204 ? null : res.json().catch(() => null)
}

// fetchMovements reads the append-only stock ledger, newest first.
// filters: { inventoryId, variantId, reason, pageSize, pageId }
async function fetchMovements ({ inventoryId, variantId, reason, pageSize = 50, pageId = 0 } = {}) {
  const params = new URLSearchParams({ page_size: pageSize, page_id: pageId })
  if (inventoryId) {params.set('inventory_id', inventoryId)}
  if (variantId) {params.set('variant_id', variantId)}
  if (reason) {params.set('reason', reason)}
  const data = await requestJSON(`${API_BASE}/stock_movements?${params}`)
  return { movements: data?.movements ?? [], total: data?.total ?? 0 }
}

export const STOCK_MOVEMENT_REASONS = [
  'purchase',
  'sale',
  'return',
  'transfer_out',
  'transfer_in',
  'adjustment',
  'count',
]

export function useStockMovements () {
  return { fetchMovements }
}
