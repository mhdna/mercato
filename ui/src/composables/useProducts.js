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

// fetchProduct returns { product, attributes, variants } for the edit dialog.
async function fetchProduct (id) {
  return requestJSON(`${API_BASE}/products/${encodeURIComponent(id)}`)
}

// updateProduct edits the shared product fields + attributes (PUT /products).
// `attributes` is the { category, subcategory, brand, ... } map the API expects.
async function updateProduct ({ id, code, name, description, is_active, attributes }) {
  return requestJSON(`${API_BASE}/products`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, code, name, description, is_active, attributes }),
  })
}

async function deleteProduct (id) {
  await requestJSON(`${API_BASE}/products/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

async function createVariant ({ product_id, color_id, size_id, price, barcode }) {
  const data = await requestJSON(`${API_BASE}/product_variants`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      product_id,
      color_id: color_id || null,
      size_id: size_id || null,
      price: price ?? null,
      barcode: barcode || null,
    }),
  })
  return data?.variant ?? data
}

async function updateVariant ({ id, color_id, size_id, barcode, price, is_active }) {
  const data = await requestJSON(`${API_BASE}/product_variants`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id,
      color_id: color_id || null,
      size_id: size_id || null,
      barcode,
      price: price ?? null,
      is_active,
    }),
  })
  return data?.variant ?? data
}

async function deleteVariant (id) {
  await requestJSON(`${API_BASE}/product_variants/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function useProducts () {
  return {
    fetchProduct,
    updateProduct,
    deleteProduct,
    createVariant,
    updateVariant,
    deleteVariant,
  }
}
