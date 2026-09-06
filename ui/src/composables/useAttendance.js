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

// Attendance punches (branch_attendance_events) normally arrive from
// kashi-pos's sync outbox on their own -- these calls are the admin-side
// view plus the manual fixes: import a CSV for a branch with no biometric
// device, and edit / delete an individual mislabelled row.
export function useAttendance () {
  async function listEvents (params = {}) {
    const url = new URL(`${API_BASE}/branch_attendance_events`)
    url.searchParams.set('page_size', params.pageSize ?? 20)
    url.searchParams.set('page_id', params.pageId ?? 0)
    if (params.branchId) {url.searchParams.set('branch_id', params.branchId)}
    const data = await requestJSON(url.toString())
    return {
      events: data?.branch_attendance_events ?? [],
      total: data?.total ?? 0,
    }
  }

  // Multipart: let the browser set the boundary Content-Type itself.
  async function importCsv (file, branchId) {
    const body = new FormData()
    body.append('branch_id', String(branchId))
    body.append('file', file)
    return requestJSON(`${API_BASE}/branch_attendance_events/import`, { method: 'POST', body })
  }

  async function updateEvent (id, payload) {
    const data = await requestJSON(`${API_BASE}/branch_attendance_events/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.branch_attendance_event ?? null
  }

  async function deleteEvent (id) {
    await requestJSON(`${API_BASE}/branch_attendance_events/${id}`, { method: 'DELETE' })
  }

  return { listEvents, importCsv, updateEvent, deleteEvent }
}
