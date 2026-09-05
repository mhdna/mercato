import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Assets are listed straight through ServerSideTable (GET /assets/), so this
// composable only wraps the write endpoints. Mirrors the request/response
// convention used across the management pages.
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

function createAsset ({ name, code, categoryId, boughtAt }) {
  return requestJSON(`${API_BASE}/assets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, code, category_id: categoryId, bought_at: boughtAt || undefined }),
  })
}

function updateAsset ({ id, name, code, categoryId, boughtAt }) {
  return requestJSON(`${API_BASE}/assets`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, name, code, category_id: categoryId, bought_at: boughtAt || undefined }),
  })
}

function deleteAsset (id) {
  return requestJSON(`${API_BASE}/assets/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function useAssets () {
  return { createAsset, updateAsset, deleteAsset }
}
