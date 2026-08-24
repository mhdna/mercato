import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

const products = ref([])
const loaded = ref(false)

async function fetchProducts (force = false) {
  if (loaded.value && !force) {
    return products.value
  }
  const res = await authFetch(`${API_BASE}/products?page_size=100&page_id=0`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  const data = await res.json()
  products.value = data.products ?? []
  loaded.value = true
  return products.value
}

export function useProducts () {
  return { products, fetchProducts }
}
