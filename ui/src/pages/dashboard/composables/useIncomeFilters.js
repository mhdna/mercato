import { computed, ref } from 'vue'

// Module-level (not per-component) so the top filter bar (BranchMenu, the
// period + scope toggles in dashboard.vue) and every Overview widget that
// reads them share one selection without prop drilling through
// dashboard.vue. null branch means "All Branches".
const selectedBranchId = ref(null)

// Overview-tab only (the Income heatmap keeps its own year selector).
// 'this-month' | '30d' | 'last-month' | 'this-year'
const period = ref('this-month')

// Which data feeds the cards/charts/tables:
// 'all' (branch + admin combined) | 'branch' | 'admin'
const scope = ref('all')

function ymd (date) {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

// { from, to } as inclusive YYYY-MM-DD local dates; the API treats `to` as
// the last day to include.
const dateRange = computed(() => {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  switch (period.value) {
    case '30d': {
      const from = new Date(today)
      from.setDate(from.getDate() - 29)
      return { from: ymd(from), to: ymd(today) }
    }
    case 'last-month': {
      const from = new Date(today.getFullYear(), today.getMonth() - 1, 1)
      const to = new Date(today.getFullYear(), today.getMonth(), 0)
      return { from: ymd(from), to: ymd(to) }
    }
    case 'this-year': {
      const from = new Date(today.getFullYear(), 0, 1)
      return { from: ymd(from), to: ymd(today) }
    }
    default: {
      // 'this-month'
      const from = new Date(today.getFullYear(), today.getMonth(), 1)
      return { from: ymd(from), to: ymd(today) }
    }
  }
})

// The period immediately before `dateRange`, same length, ending the day
// before it starts. A partial "this month" compares against the equivalent
// span of last month, which is the fair like-for-like for a running period.
const previousDateRange = computed(() => {
  const from = new Date(`${dateRange.value.from}T00:00:00`)
  const to = new Date(`${dateRange.value.to}T00:00:00`)
  const spanDays = Math.round((to - from) / 86_400_000) + 1
  const prevTo = new Date(from)
  prevTo.setDate(prevTo.getDate() - 1)
  const prevFrom = new Date(prevTo)
  prevFrom.setDate(prevFrom.getDate() - (spanDays - 1))
  return { from: ymd(prevFrom), to: ymd(prevTo) }
})

// Query string for the /dashboard/* endpoints, reflecting all three
// filters. Reactive callers should read `period.value` / `scope.value` /
// `selectedBranchId.value` in a watcher and rebuild the URL. Pass `range`
// to override the date window (e.g. previousDateRange for a comparison).
function dashboardQuery (range = dateRange.value) {
  const params = new URLSearchParams({
    scope: scope.value,
    date_from: range.from,
    date_to: range.to,
  })
  if (selectedBranchId.value != null) {
    params.set('branch_id', String(selectedBranchId.value))
  }
  return params.toString()
}

export function useIncomeFilters () {
  return { selectedBranchId, period, scope, dateRange, previousDateRange, dashboardQuery }
}
