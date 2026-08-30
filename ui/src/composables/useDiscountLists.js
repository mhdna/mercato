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

const discountLists = ref([])
const loaded = ref(false)

async function fetchDiscountLists (force = false) {
  if (loaded.value && !force) {
    return discountLists.value
  }
  const data = await requestJSON(`${API_BASE}/discount_lists?page_size=100&page_id=0`)
  discountLists.value = Array.isArray(data) ? data : (data?.discount_lists ?? [])
  loaded.value = true
  return discountLists.value
}

async function createDiscountList (payload) {
  const item = await requestJSON(`${API_BASE}/discount_lists`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  await fetchDiscountLists(true)
  return item
}

async function updateDiscountList (payload) {
  await requestJSON(`${API_BASE}/discount_lists`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  await fetchDiscountLists(true)
}

async function deleteDiscountList (id) {
  await requestJSON(`${API_BASE}/discount_lists/${id}`, { method: 'DELETE' })
  await fetchDiscountLists(true)
}

async function fetchDiscountListItems (discountListId) {
  const data = await requestJSON(`${API_BASE}/discount_lists/${discountListId}/items`)
  // API returns { items: [...] }; older callers expect a bare array.
  return Array.isArray(data) ? data : (data?.items ?? [])
}

// Upsert: POST does INSERT ... ON CONFLICT DO UPDATE, so this adds or edits.
async function saveDiscountListItem ({ discountListId, productId, discount }) {
  return requestJSON(`${API_BASE}/discount_lists/items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ discount_list_id: discountListId, product_id: productId, discount }),
  })
}

async function deleteDiscountListItem (discountListId, productId) {
  await requestJSON(`${API_BASE}/discount_lists/${discountListId}/items/${productId}`, { method: 'DELETE' })
}

async function fetchDiscountListBranchIds (discountListId) {
  const data = await requestJSON(`${API_BASE}/discount_lists/${discountListId}/branches`)
  return data?.branch_ids ?? []
}

async function setDiscountListBranches (discountListId, branchIds) {
  return requestJSON(`${API_BASE}/discount_lists/${discountListId}/branches`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ branch_ids: branchIds }),
  })
}

export function useDiscountLists () {
  return {
    discountLists,
    fetchDiscountLists,
    createDiscountList,
    updateDiscountList,
    deleteDiscountList,
    fetchDiscountListItems,
    saveDiscountListItem,
    deleteDiscountListItem,
    fetchDiscountListBranchIds,
    setDiscountListBranches,
  }
}
