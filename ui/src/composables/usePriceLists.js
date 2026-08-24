import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

const priceLists = ref([])
const loaded = ref(false)

async function fetchPriceLists (force = false) {
  if (loaded.value && !force) {
    return priceLists.value
  }
  const res = await authFetch(`${API_BASE}/price_lists?page_size=10&page_id=0`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  const data = await res.json()
  priceLists.value = Array.isArray(data) ? data : (data.price_lists ?? [])
  loaded.value = true
  return priceLists.value
}

export function usePriceLists () {
  return { priceLists, fetchPriceLists }
}
