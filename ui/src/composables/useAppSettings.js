import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

// Global (not per-user/per-branch) app settings, backed by the single-row
// app_settings table -- see db/migrations/000050_create_app_settings and
// api/app_settings.go. Consumed by stores/settings.js; prefer that store
// over calling this composable directly.
export function useAppSettings () {
  async function getAppSettings () {
    const data = await requestJSON(`${API_BASE}/app_settings`)
    return data?.settings ?? null
  }

  // Replaces the whole settings row -- the caller sends the full object
  // (all fields), not a partial patch.
  async function updateAppSettings (payload) {
    const data = await requestJSON(`${API_BASE}/app_settings`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.settings ?? null
  }

  return { getAppSettings, updateAppSettings }
}
