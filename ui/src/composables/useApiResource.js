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

/**
 * Shared module-level cache + CRUD helpers for a simple REST resource,
 * mirroring the pattern in useColorsAndSizes.js. Meant for reference/lookup
 * data used across forms and for simple list+create+edit+delete pages.
 *
 * `path` is the base used for POST/PUT/DELETE (no trailing slash, e.g. "/clients").
 * `listPath` defaults to `path + '/'` since most list routes are registered with
 * a trailing slash in this API; pass it explicitly for the handful of entities
 * whose list route has no trailing slash (e.g. shifts, coupons, transfers).
 */
export function createResource ({ path, listPath, rootKey, idKey = 'id' }) {
  const items = ref([])
  const loaded = ref(false)
  const resolvedListPath = listPath ?? `${path}/`

  async function fetchAll (force = false) {
    if (loaded.value && !force) {
      return items.value
    }
    const data = await requestJSON(`${API_BASE}${resolvedListPath}`)
    items.value = Array.isArray(data) ? data : (data?.[rootKey] ?? [])
    loaded.value = true
    return items.value
  }

  async function create (body) {
    const item = await requestJSON(`${API_BASE}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    items.value.push(item)
    return item
  }

  async function update (body) {
    const item = await requestJSON(`${API_BASE}${path}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const idx = items.value.findIndex(i => i[idKey] === body[idKey])
    if (idx !== -1) {
      items.value[idx] = item ?? { ...items.value[idx], ...body }
    }
    return item
  }

  async function remove (id) {
    await requestJSON(`${API_BASE}${path}/${encodeURIComponent(id)}`, { method: 'DELETE' })
    items.value = items.value.filter(i => i[idKey] !== id)
  }

  return { items, loaded, fetchAll, create, update, remove, requestJSON }
}
