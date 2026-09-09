import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/purchases', listPath: '/purchases', rootKey: 'purchases' })

// createPurchase creates a draft purchase invoice with its line items.
// items: [{ variantId, quantity, unitPrice }]
async function createPurchase ({ supplierId, inventoryId, code = '', currencyCode, purchasedAt, note = '', items = [] }) {
  return resource.requestJSON(`${API_BASE}/purchases`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      supplier_id: supplierId,
      inventory_id: inventoryId,
      code,
      currency_code: currencyCode,
      purchased_at: purchasedAt || undefined,
      note,
      items: items.map(i => ({ variant_id: i.variantId, quantity: i.quantity, unit_price: i.unitPrice })),
    }),
  })
}

async function fetchPurchase (id) {
  return resource.requestJSON(`${API_BASE}/purchases/${id}`)
}

async function addPurchaseItem ({ purchaseId, variantId, quantity, unitPrice, currencyCode }) {
  return resource.requestJSON(`${API_BASE}/purchase_items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      purchase_id: purchaseId,
      variant_id: variantId,
      quantity,
      unit_price: unitPrice,
      currency_code: currencyCode,
    }),
  })
}

async function receivePurchase (id) {
  return resource.requestJSON(`${API_BASE}/purchases/${id}/receive`, { method: 'POST' })
}

async function cancelPurchase (id) {
  return resource.requestJSON(`${API_BASE}/purchases/${id}/cancel`, { method: 'POST' })
}

export function usePurchases () {
  return {
    purchases: resource.items,
    fetchPurchases: resource.fetchAll,
    createPurchase,
    fetchPurchase,
    addPurchaseItem,
    receivePurchase,
    cancelPurchase,
  }
}
