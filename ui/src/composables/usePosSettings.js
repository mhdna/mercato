import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// GET /pos_settings -- read-only reference data scoped to one cashbox,
// already used by kashi-pos's own dropdowns (salespersons/cashbox
// accounts/currencies). Reused here to populate the same salesperson
// picker on kashi's own local invoice-create form.
export function usePosSettings () {
  async function getPosSettings (cashboxId) {
    const res = await authFetch(`${API_BASE}/pos_settings?cashbox_id=${cashboxId}`)
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.error || `Request failed with status ${res.status}`)
    }
    return res.json().catch(() => null)
  }

  return { getPosSettings }
}
