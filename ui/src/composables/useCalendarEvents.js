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

// User-defined one-off calendar events (see api/calendar_event.go). These
// are merged with the built-in retail calendar by getEventsForRange() in
// data/retailCalendarEvents.js -- callers pass the rows from
// listCalendarEvents() straight through as its `custom` option.
export function useCalendarEvents () {
  async function listCalendarEvents () {
    const data = await requestJSON(`${API_BASE}/calendar_events`)
    return data?.events ?? []
  }

  async function createCalendarEvent (payload) {
    const data = await requestJSON(`${API_BASE}/calendar_events`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.event ?? null
  }

  async function deleteCalendarEvent (id) {
    await requestJSON(`${API_BASE}/calendar_events/${id}`, { method: 'DELETE' })
  }

  return { listCalendarEvents, createCalendarEvent, deleteCalendarEvent }
}
