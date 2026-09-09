import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url) {
  const res = await authFetch(url)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

// Overall inventory health: days-of-inventory, value, and the same
// breakdown per inventory location -- see api/stock_health.go. Backs both
// the Stock Health page and the sidebar status dot.
export function useStockHealth () {
  async function fetchStockHealth () {
    return requestJSON(`${API_BASE}/stock_health`)
  }

  return { fetchStockHealth }
}
