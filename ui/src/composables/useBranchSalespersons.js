import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Same shape as useBranchTargets.js: list/create are nested under a branch
// id (/branches/:id/salespersons), while update/deactivate address the
// salesperson directly (/salespersons/:id[/active]).

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

export function useBranchSalespersons () {
  async function listBranchSalespersons (branchId) {
    const data = await requestJSON(`${API_BASE}/branches/${branchId}/salespersons`)
    return data?.salespersons ?? []
  }

  async function createBranchSalesperson (branchId, name) {
    return requestJSON(`${API_BASE}/branches/${branchId}/salespersons`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
  }

  async function updateBranchSalesperson (id, name) {
    return requestJSON(`${API_BASE}/salespersons/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
  }

  async function setBranchSalespersonActive (id, isActive) {
    return requestJSON(`${API_BASE}/salespersons/${id}/active`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ is_active: isActive }),
    })
  }

  return {
    listBranchSalespersons,
    createBranchSalesperson,
    updateBranchSalesperson,
    setBranchSalespersonActive,
  }
}
