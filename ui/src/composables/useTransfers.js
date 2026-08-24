import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/transfers', listPath: '/transfers', rootKey: 'transfers' })

async function fetchTransferItems (transferId) {
  const data = await resource.requestJSON(`${API_BASE}/transfer_items/${transferId}`)
  return data ?? []
}

async function createTransferItem ({ transferId, productId, quantity }) {
  return resource.requestJSON(`${API_BASE}/transfer_items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ transfer_id: transferId, product_id: productId, quantity }),
  })
}

export function useTransfers () {
  return {
    transfers: resource.items,
    fetchTransfers: resource.fetchAll,
    createTransfer: resource.create,
    updateTransfer: resource.update,
    fetchTransferItems,
    createTransferItem,
  }
}
