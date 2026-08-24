import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Branches don't fit the generic createResource shape (see useApiResource.js):
// create returns { branch, api_key } instead of the bare row, there's no
// update/delete, and activate/deactivate/rotate_key are bespoke POST actions.
// So this composable is hand-written rather than built on createResource.

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

const branches = ref([])
const loaded = ref(false)

export function useBranches () {
  async function fetchBranches (force = false) {
    if (loaded.value && !force) {
      return branches.value
    }
    const data = await requestJSON(`${API_BASE}/branches`)
    branches.value = data?.branches ?? []
    loaded.value = true
    return branches.value
  }

  async function createBranch ({ name, code }) {
    const data = await requestJSON(`${API_BASE}/branches`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, code }),
    })
    branches.value.push(data.branch)
    return data // { branch, api_key } -- api_key is shown once, never again
  }

  async function setBranchActive (id, active) {
    await requestJSON(`${API_BASE}/branches/${id}/${active ? 'activate' : 'deactivate'}`, { method: 'POST' })
    const branch = branches.value.find(b => b.id === id)
    if (branch) {
      branch.is_active = active
    }
  }

  async function rotateBranchKey (id) {
    return requestJSON(`${API_BASE}/branches/${id}/rotate_key`, { method: 'POST' }) // { id, api_key }
  }

  return { branches, fetchBranches, createBranch, setBranchActive, rotateBranchKey }
}
