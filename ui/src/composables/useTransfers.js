import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/transfers', listPath: '/transfers', rootKey: 'transfers' })

// createTransfer creates a draft transfer with its line items in one call.
// items: [{ variantId, quantity }]
async function createTransfer ({ fromInventoryId, toInventoryId, note = '', items = [] }) {
  return resource.requestJSON(`${API_BASE}/transfers`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      from_inventory_id: fromInventoryId,
      to_inventory_id: toInventoryId,
      note,
      items: items.map(i => ({ variant_id: i.variantId, quantity: i.quantity })),
    }),
  })
}

async function fetchTransfer (id) {
  return resource.requestJSON(`${API_BASE}/transfers/${id}`)
}

async function fetchTransferItems (transferId) {
  const data = await resource.requestJSON(`${API_BASE}/transfer_items/${transferId}`)
  return data ?? []
}

async function createTransferItem ({ transferId, variantId, quantity }) {
  return resource.requestJSON(`${API_BASE}/transfer_items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ transfer_id: transferId, variant_id: variantId, quantity }),
  })
}

const stage = action => id => resource.requestJSON(`${API_BASE}/transfers/${id}/${action}`, { method: 'POST' })

export function useTransfers () {
  return {
    transfers: resource.items,
    fetchTransfers: resource.fetchAll,
    createTransfer,
    updateTransfer: resource.update,
    fetchTransfer,
    fetchTransferItems,
    createTransferItem,
    dispatchTransfer: stage('dispatch'),
    receiveTransfer: stage('receive'),
  }
}
