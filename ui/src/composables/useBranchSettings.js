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

export function useBranchSettings () {
  // Returns null (not an error) if this branch hasn't pushed settings yet
  // -- a fresh branch that's never synced is an expected, normal state.
  async function getBranchSettings (branchId) {
    const res = await authFetch(`${API_BASE}/branches/${branchId}/settings`)
    if (res.status === 404) {
      return null
    }
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.error || `Request failed with status ${res.status}`)
    }
    const data = await res.json().catch(() => null)
    return data?.settings ?? null
  }

  // Enqueues a remote "update_settings" command -- see api/branch_command.go
  // and kashi-pos's commands.go. Returns immediately once kashi has queued
  // it; it does NOT wait for the branch to execute it (that happens
  // asynchronously, usually within a second or two if the branch is
  // online, via the WebSocket push, or on its next sync tick otherwise).
  async function updateBranchSettings (branchId, payload) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/commands`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ type: 'update_settings', payload }),
    })
    return data?.command ?? null
  }

  async function listBranchCommands (branchId) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/commands?page_size=10&page_id=0`)
    return data?.branch_commands ?? []
  }

  // Records which settings groups the admin has handed back to the branch
  // (e.g. 'secrets', 'device_ids'). Those groups are shown read-only and
  // never included in an update_settings command payload. Applies
  // immediately -- it's central admin state, not a branch command.
  async function setBranchSettingsManagedLocally (branchId, groups) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/settings/managed_locally`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ managed_locally: groups }),
    })
    return data?.settings ?? null
  }

  return { getBranchSettings, updateBranchSettings, listBranchCommands, setBranchSettingsManagedLocally }
}
