import { API_BASE } from '@/config'
import { authFetch } from './useApi'

export function useDailyIncome () {
  async function getDailyIncome (year, branchId) {
    const url = new URL(`${API_BASE}/branch_invoices/daily_income`, window.location.origin)
    url.searchParams.set('year', year)
    if (branchId) {
      url.searchParams.set('branch_id', branchId)
    }

    const res = await authFetch(url.toString())
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.error || `Request failed with status ${res.status}`)
    }
    const data = await res.json().catch(() => null)
    return data?.days ?? []
  }

  return { getDailyIncome }
}
