import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

async function requestJSON (url) {
  const res = await authFetch(url)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

// A SKU label like "PRD001 Shirt · Blue / M · 1234567890".
export function variantLabel (v) {
  const bits = [v.product_code, v.product_name].filter(Boolean).join(' ')
  const cs = [v.color_name, v.size_name].filter(Boolean).join(' / ')
  return [bits, cs, v.barcode].filter(Boolean).join(' · ')
}

// useVariants backs server-side-searched SKU pickers (v-autocomplete).
// Results come from GET /barcodes, which returns variant rows joined to
// their product + colour/size.
export function useVariants () {
  const variants = ref([])
  const loading = ref(false)

  async function searchVariants (search = '') {
    loading.value = true
    try {
      const params = new URLSearchParams({ page_id: '0', page_size: '50' })
      if (search) {params.set('search', search)}
      const data = await requestJSON(`${API_BASE}/barcodes?${params}`)
      variants.value = (data?.variants ?? []).map(v => ({ ...v, label: variantLabel(v) }))
      return variants.value
    } finally {
      loading.value = false
    }
  }

  return { variants, loading, searchVariants }
}
