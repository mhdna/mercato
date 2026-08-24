import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/purchases', listPath: '/purchases', rootKey: 'purchases' })

async function addPurchaseItem ({ purchaseId, productId, quantity, unitPrice, currencyCode }) {
  return resource.requestJSON(`${API_BASE}/purchases/items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      purchase_id: purchaseId,
      product_id: productId,
      quantity,
      unit_price: unitPrice,
      currency_code: currencyCode,
    }),
  })
}

export function usePurchases () {
  return {
    purchases: resource.items,
    fetchPurchases: resource.fetchAll,
    createPurchase: resource.create,
    addPurchaseItem,
  }
}
