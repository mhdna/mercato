import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Branch-reported visitor-counter presses, already rolled up server-side to
// one row per branch per local day (see api/branch_visitor.go's
// listBranchVisitorDays). There's no create path here -- rows only ever
// arrive via branch sync -- so this is a read-only helper, same shape as
// useBranchTargets' list calls.

async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed (${res.status})`)
  }
  return res.json()
}

export function useBranchVisitors () {
  // listVisitorDays returns the most recent { branch_id, day, in_count,
  // out_count } rows, newest day first. pageSize is capped at 100 by the
  // API's page_size validation.
  async function listVisitorDays ({ branchId, pageSize = 100 } = {}) {
    const url = new URL(`${API_BASE}/branch_visitor_events`)
    url.searchParams.set('page_size', String(Math.min(pageSize, 100)))
    url.searchParams.set('page_id', '0')
    if (branchId) url.searchParams.set('branch_id', String(branchId))
    const data = await requestJSON(url.toString())
    return data?.visitor_days ?? []
  }

  return { listVisitorDays }
}
