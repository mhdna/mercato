import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Targets don't fit the generic createResource shape (see useApiResource.js):
// the list/create routes are nested under a branch id (/branches/:id/targets)
// while update/delete address the target directly (/branch_targets/:id), and
// there's no single shared cache since each branch's dialog wants its own list.

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

// <input type="date"> gives plain "YYYY-MM-DD" strings, but the API's
// date_from/date_to bind as Go time.Time (RFC3339 only) -- widen to
// midnight UTC before sending.
function toRFC3339 (dateOnly) {
  return `${dateOnly}T00:00:00Z`
}

export function useBranchTargets () {
  async function listBranchTargets (branchId) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/targets`)
    return data?.targets ?? []
  }

  // `color` is optional in both -- the API auto-assigns one from a fixed
  // palette on create if left blank, and leaves the existing color alone
  // on update if left blank.
  async function createBranchTarget (branchId, { dateFrom, dateTo, targetAmount, color }) {
    return requestJSON(`${API_BASE}/branches/${branchId}/targets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        date_from: toRFC3339(dateFrom),
        date_to: toRFC3339(dateTo),
        target_amount: targetAmount,
        color: color || undefined,
      }),
    })
  }

  async function updateBranchTarget (id, { dateFrom, dateTo, targetAmount, color }) {
    return requestJSON(`${API_BASE}/branch_targets/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        date_from: toRFC3339(dateFrom),
        date_to: toRFC3339(dateTo),
        target_amount: targetAmount,
        color: color || undefined,
      }),
    })
  }

  async function deleteBranchTarget (id) {
    await requestJSON(`${API_BASE}/branch_targets/${id}`, { method: 'DELETE' })
  }

  async function listBranchTargetSeries (branchId) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/target_series`)
    return data?.series ?? []
  }

  // startDay: day-of-month a period starts on (1-31, clamped server-side
  // in short months). intervalCount: how many months each period spans.
  async function createBranchTargetSeries (branchId, { targetAmount, startDay, intervalCount, color }) {
    return requestJSON(`${API_BASE}/branches/${branchId}/target_series`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        target_amount: targetAmount,
        start_day: startDay,
        interval_count: intervalCount,
        color: color || undefined,
      }),
    })
  }

  async function setBranchTargetSeriesActive (id, active) {
    return requestJSON(`${API_BASE}/target_series/${id}/active`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ active }),
    })
  }

  return {
    listBranchTargets,
    createBranchTarget,
    updateBranchTarget,
    deleteBranchTarget,
    listBranchTargetSeries,
    createBranchTargetSeries,
    setBranchTargetSeriesActive,
  }
}
