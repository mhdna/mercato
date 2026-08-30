import { ref, watch } from 'vue'
import { API_BASE } from '@/config'
import { useIncomeFilters } from '@/pages/dashboard/composables/useIncomeFilters'
import { authFetch } from './useApi'
import { useDashboardLiveEvents } from './useDashboardLiveEvents'

// Feeds the Overview tab's summary cards and trend charts from
// GET /dashboard/summary -- fetched twice, once for the selected period and
// once for the one before it (prevSummary, for period-over-period deltas) --
// re-fetching whenever the shared period / scope / branch filters change.
export function useDashboardSummary () {
  const { period, scope, selectedBranchId, previousDateRange, dashboardQuery } = useIncomeFilters()

  const summary = ref(null)
  // Same payload shape for the immediately-preceding period, so widgets can
  // show period-over-period deltas without a second composable.
  const prevSummary = ref(null)
  const loading = ref(false)
  const error = ref('')

  let reqToken = 0

  async function fetchSummary (query) {
    const res = await authFetch(`${API_BASE}/dashboard/summary?${query}`)
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body.error || `Request failed with status ${res.status}`)
    }
    return res.json()
  }

  async function refresh () {
    const token = ++reqToken
    loading.value = true
    error.value = ''
    try {
      const [data, prev] = await Promise.all([
        fetchSummary(dashboardQuery()),
        fetchSummary(dashboardQuery(previousDateRange.value)),
      ])
      if (token !== reqToken) {
        return
      } // a newer request already superseded this one
      summary.value = data
      prevSummary.value = prev
    } catch (error_) {
      if (token !== reqToken) {
        return
      }
      error.value = error_.message
      // Keep the last good numbers on screen -- a failed background refresh
      // shouldn't blank every card. Only clear if we never had data.
    } finally {
      if (token === reqToken) {
        loading.value = false
      }
    }
  }

  watch([period, scope, selectedBranchId], refresh, { immediate: true })

  // Re-pull when a branch syncs new data.
  useDashboardLiveEvents(refresh)

  return { summary, prevSummary, loading, error, refresh }
}
