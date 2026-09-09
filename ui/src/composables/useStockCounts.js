import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/stock_counts', listPath: '/stock_counts', rootKey: 'stock_counts' })

// items: [{ variantId, countedQuantity }]
function toItemsPayload (items) {
  return items.map(i => ({ variant_id: i.variantId, counted_quantity: i.countedQuantity }))
}

async function createStockCount ({ inventoryId, note = '', items = [] }) {
  return resource.requestJSON(`${API_BASE}/stock_counts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ inventory_id: inventoryId, note, items: toItemsPayload(items) }),
  })
}

async function updateStockCount ({ id, inventoryId, note = '', items = [] }) {
  return resource.requestJSON(`${API_BASE}/stock_counts`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, inventory_id: inventoryId, note, items: toItemsPayload(items) }),
  })
}

async function fetchStockCount (id) {
  return resource.requestJSON(`${API_BASE}/stock_counts/${id}`)
}

async function postStockCount (id) {
  return resource.requestJSON(`${API_BASE}/stock_counts/${id}/post`, { method: 'POST' })
}

async function cancelStockCount (id) {
  return resource.requestJSON(`${API_BASE}/stock_counts/${id}/cancel`, { method: 'POST' })
}

// The count sheet seeds a new draft: every SKU in the inventory with its
// live on-hand. Reuses the same shape as GET /inventories/:id/stock.
//
// Paged in batches rather than one large request: a big inventory can hold
// thousands of SKUs, and a single multi-thousand-row response both blocks
// the UI on one huge fetch and used to arrive with a hard 500-row cap
// (silently incomplete). `onPage` is called with each batch as it arrives
// so the caller can render progressively instead of waiting for the whole
// inventory.
const COUNT_SHEET_PAGE_SIZE = 200

async function fetchCountSheet (inventoryId, { search = '', onPage } = {}) {
  const all = []
  let pageId = 0
  for (;;) {
    const params = new URLSearchParams({ page_id: String(pageId), page_size: String(COUNT_SHEET_PAGE_SIZE) })
    if (search) params.set('search', search)
    const data = await resource.requestJSON(`${API_BASE}/inventories/${inventoryId}/count_sheet?${params}`)
    const rows = data?.stock ?? []
    all.push(...rows)
    onPage?.(rows, { loaded: all.length, total: data?.total ?? all.length })
    if (rows.length < COUNT_SHEET_PAGE_SIZE || all.length >= (data?.total ?? 0)) break
    pageId += COUNT_SHEET_PAGE_SIZE
  }
  return all
}

export function useStockCounts () {
  return {
    stockCounts: resource.items,
    fetchStockCounts: resource.fetchAll,
    createStockCount,
    updateStockCount,
    fetchStockCount,
    postStockCount,
    cancelStockCount,
    fetchCountSheet,
  }
}
