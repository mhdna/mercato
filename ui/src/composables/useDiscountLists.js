import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/discount_lists', listPath: '/discount_lists', rootKey: 'discount_lists' })

async function fetchDiscountListItems (discountListId) {
  const data = await resource.requestJSON(`${API_BASE}/discount_lists/${discountListId}/items`)
  return data ?? []
}

async function createDiscountListItem ({ discountListId, productId, discount }) {
  return resource.requestJSON(`${API_BASE}/discount_lists/items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ discount_list_id: discountListId, product_id: productId, discount }),
  })
}

async function deleteDiscountListItem (discountListId, productId) {
  await resource.requestJSON(`${API_BASE}/discount_lists/${discountListId}/items/${productId}`, { method: 'DELETE' })
}

export function useDiscountLists () {
  return {
    discountLists: resource.items,
    fetchDiscountLists: resource.fetchAll,
    createDiscountList: resource.create,
    updateDiscountList: resource.update,
    fetchDiscountListItems,
    createDiscountListItem,
    deleteDiscountListItem,
  }
}
