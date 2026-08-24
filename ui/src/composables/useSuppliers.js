import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

const suppliers = ref([])
const loaded = ref(false)

async function fetchSuppliers (force = false) {
  if (loaded.value && !force) {
    return suppliers.value
  }
  const res = await authFetch(`${API_BASE}/suppliers?page_size=100&page_id=0`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  const data = await res.json()
  suppliers.value = Array.isArray(data) ? data : (data.suppliers ?? [])
  loaded.value = true
  return suppliers.value
}

export function useSuppliers () {
  return { suppliers, fetchSuppliers }
}
