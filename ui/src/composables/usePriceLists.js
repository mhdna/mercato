import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  if (res.status === 204) {
    return null
  }
  return res.json().catch(() => null)
}

const priceLists = ref([])
const loaded = ref(false)

async function fetchPriceLists (force = false) {
  if (loaded.value && !force) {
    return priceLists.value
  }
  const data = await requestJSON(`${API_BASE}/price_lists?page_size=100&page_id=0`)
  priceLists.value = Array.isArray(data) ? data : (data?.price_lists ?? [])
  loaded.value = true
  return priceLists.value
}

async function createPriceList (payload) {
  const item = await requestJSON(`${API_BASE}/price_lists`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  await fetchPriceLists(true)
  return item
}

async function updatePriceList (payload) {
  await requestJSON(`${API_BASE}/price_lists`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  await fetchPriceLists(true)
}

async function deletePriceList (id) {
  await requestJSON(`${API_BASE}/price_lists/${id}`, { method: 'DELETE' })
  await fetchPriceLists(true)
}

async function fetchPriceListItems (priceListId) {
  const data = await requestJSON(`${API_BASE}/price_lists/${priceListId}/items`)
  return data?.items ?? []
}

// Upsert -- the API's POST /price_lists/items does INSERT ... ON CONFLICT
// DO UPDATE, so this both adds a new custom price and edits an existing one.
async function savePriceListItem ({ priceListId, productId, price }) {
  return requestJSON(`${API_BASE}/price_lists/items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ price_list_id: priceListId, product_id: productId, price }),
  })
}

async function deletePriceListItem (priceListId, productId) {
  await requestJSON(`${API_BASE}/price_lists/${priceListId}/items/${productId}`, { method: 'DELETE' })
}

async function fetchPriceListBranchIds (priceListId) {
  const data = await requestJSON(`${API_BASE}/price_lists/${priceListId}/branches`)
  return data?.branch_ids ?? []
}

async function setPriceListBranches (priceListId, branchIds) {
  return requestJSON(`${API_BASE}/price_lists/${priceListId}/branches`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ branch_ids: branchIds }),
  })
}

export function usePriceLists () {
  return {
    priceLists,
    fetchPriceLists,
    createPriceList,
    updatePriceList,
    deletePriceList,
    fetchPriceListItems,
    savePriceListItem,
    deletePriceListItem,
    fetchPriceListBranchIds,
    setPriceListBranches,
  }
}
