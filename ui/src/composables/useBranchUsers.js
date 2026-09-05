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

// Branch POS users: an admin-managed, branch-scoped roster mirrored to the
// till read-only. The PIN is never returned here -- setPin enqueues a
// set_branch_user_pin command the till hashes locally.
export function useBranchUsers () {
  async function listBranchUsers (branchId) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/branch_users`)
    return data?.branch_users ?? []
  }

  async function createBranchUser (branchId, payload) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/branch_users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.branch_user ?? null
  }

  async function updateBranchUser (id, payload) {
    const data = await requestJSON(`${API_BASE}/branch_users/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return data?.branch_user ?? null
  }

  async function setBranchUserPin (id, pin) {
    await requestJSON(`${API_BASE}/branch_users/${id}/pin`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pin }),
    })
  }

  async function deleteBranchUser (id) {
    await requestJSON(`${API_BASE}/branch_users/${id}`, { method: 'DELETE' })
  }

  return { listBranchUsers, createBranchUser, updateBranchUser, setBranchUserPin, deleteBranchUser }
}
