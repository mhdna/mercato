import { ref } from 'vue'
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

/**
 * Shared store for the "list of lists, each with per-product line items and a
 * branch assignment" shape — price lists and discount lists are the same
 * entity with one differently-named value column (`price` cents vs
 * `discount` percent). `base` is the REST root ("/price_lists"), `rootKey`
 * the list envelope key ("price_lists"), `valueKey` the line-item column.
 *
 * The list of lists is small (capped page) and module-cached; line items are
 * paged server-side by the caller (ServerSideTable), so this exposes an
 * items *URL builder* rather than fetching them here.
 */
export function useValueListStore ({ base, rootKey, listIdKey, valueKey }) {
  const list = ref([])
  const loaded = ref(false)

  async function fetchLists (force = false) {
    if (loaded.value && !force) {
      return list.value
    }
    const data = await requestJSON(`${API_BASE}${base}?page_size=100&page_id=0`)
    list.value = Array.isArray(data) ? data : (data?.[rootKey] ?? [])
    loaded.value = true
    return list.value
  }

  async function createList (payload) {
    const item = await requestJSON(`${API_BASE}${base}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    await fetchLists(true)
    return item
  }

  async function updateList (payload) {
    await requestJSON(`${API_BASE}${base}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    await fetchLists(true)
  }

  async function deleteList (id) {
    await requestJSON(`${API_BASE}${base}/${id}`, { method: 'DELETE' })
    await fetchLists(true)
  }

  // Absolute URL for ServerSideTable to page /<base>/<id>/items itself.
  const itemsURL = id => `${API_BASE}${base}/${id}/items`

  // Upsert — POST does INSERT ... ON CONFLICT DO UPDATE, so it adds or edits.
  async function saveItem (listId, productId, value) {
    return requestJSON(`${API_BASE}${base}/items`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ [listIdKey]: listId, product_id: productId, [valueKey]: value }),
    })
  }

  async function deleteItem (listId, productId) {
    await requestJSON(`${API_BASE}${base}/${listId}/items/${productId}`, { method: 'DELETE' })
  }

  async function fetchBranchIds (listId) {
    const data = await requestJSON(`${API_BASE}${base}/${listId}/branches`)
    return data?.branch_ids ?? []
  }

  async function setBranches (listId, branchIds) {
    return requestJSON(`${API_BASE}${base}/${listId}/branches`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ branch_ids: branchIds }),
    })
  }

  return {
    list, loaded, fetchLists, createList, updateList, deleteList,
    itemsURL, saveItem, deleteItem, fetchBranchIds, setBranches,
  }
}
