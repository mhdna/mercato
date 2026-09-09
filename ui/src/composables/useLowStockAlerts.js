import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  if (res.status === 204) return null
  return res.json().catch(() => null)
}

// Low stock alerts: every active product's on-hand (summed across every
// SKU and inventory) compared against its effective threshold (its own
// override, else the global default) -- see api/low_stock.go.
export function useLowStockAlerts () {
  async function fetchSummary () {
    return requestJSON(`${API_BASE}/low_stock/summary`)
  }

  // Cheap count-only fetch (page_size=1) backing the "Ignored" tab badge.
  async function fetchIgnoredCount () {
    const data = await requestJSON(`${API_BASE}/low_stock/ignored?page_size=1&page_id=0`)
    return data?.total ?? 0
  }

  async function setProductThreshold (productId, threshold) {
    return requestJSON(`${API_BASE}/products/${productId}/low_stock_threshold`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ threshold }),
    })
  }

  async function setDefaultThreshold (threshold) {
    const data = await requestJSON(`${API_BASE}/app_settings/default_low_stock_threshold`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ threshold }),
    })
    return data?.settings ?? null
  }

  // Mute ("don't care if it's out of stock") or unmute one product.
  async function setAlertsEnabled (productId, enabled) {
    return requestJSON(`${API_BASE}/products/${productId}/low_stock_alerts`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    })
  }

  return { fetchSummary, fetchIgnoredCount, setProductThreshold, setDefaultThreshold, setAlertsEnabled }
}
