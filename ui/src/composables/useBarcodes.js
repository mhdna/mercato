import { API_BASE } from '@/config'
import { authFetch } from './useApi'

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

export function useBarcodes () {
  async function listVariants ({ page = 0, pageSize = 25, search = '' } = {}) {
    const params = new URLSearchParams({
      page_id: String(page),
      page_size: String(pageSize),
    })
    if (search) {
      params.set('search', search)
    }
    const data = await requestJSON(`${API_BASE}/barcodes?${params}`)
    return { variants: data?.variants ?? [], total: data?.total ?? 0 }
  }

  // Assigns fresh unique EAN-13 barcodes to selected variants that don't
  // already have a valid one. Returns { updated, skipped }.
  async function assignBarcodes (variantIds) {
    return requestJSON(`${API_BASE}/barcodes/assign`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ variant_ids: variantIds }),
    })
  }

  // Returns a Blob of the generated label PDF.
  async function printLabels ({ items, separator, label }) {
    const res = await authFetch(`${API_BASE}/barcodes/print`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items, separator, label }),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.error || `Request failed with status ${res.status}`)
    }
    return res.blob()
  }

  return { listVariants, assignBarcodes, printLabels }
}
