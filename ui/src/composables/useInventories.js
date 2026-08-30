import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/inventories', rootKey: 'inventories' })

// fetchStock returns the per-SKU on-hand quantity + moving-average cost for
// one inventory: { stock: [...], total }.
async function fetchStock (inventoryId, { search = '', pageSize = 200, pageId = 0 } = {}) {
  const params = new URLSearchParams({ search, page_size: pageSize, page_id: pageId })
  const data = await resource.requestJSON(`${API_BASE}/inventories/${inventoryId}/stock?${params}`)
  return { stock: data?.stock ?? [], total: data?.total ?? 0 }
}

// createAdjustment applies a manual correction to one SKU in one inventory.
// mode is 'delta' (signed change) or 'count' (absolute counted on-hand).
async function createAdjustment (inventoryId, { variantId, mode, quantity, note = '' }) {
  return resource.requestJSON(`${API_BASE}/inventories/${inventoryId}/adjustments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ variant_id: variantId, mode, quantity, note }),
  })
}

export function useInventories () {
  return {
    inventories: resource.items,
    fetchInventories: resource.fetchAll,
    createInventory: resource.create,
    updateInventory: resource.update,
    deleteInventory: resource.remove,
    fetchStock,
    createAdjustment,
  }
}
