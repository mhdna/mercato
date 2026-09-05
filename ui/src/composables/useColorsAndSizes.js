import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

const colors = ref([])
const sizes = ref([])
const colorsLoaded = ref(false)
const sizesLoaded = ref(false)

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

async function fetchColors (force = false) {
  if (colorsLoaded.value && !force) {
    return colors.value
  }
  const data = await requestJSON(`${API_BASE}/colors`)
  colors.value = data.colors ?? []
  colorsLoaded.value = true
  return colors.value
}

async function fetchSizes (force = false) {
  if (sizesLoaded.value && !force) {
    return sizes.value
  }
  const data = await requestJSON(`${API_BASE}/sizes`)
  sizes.value = data.sizes ?? []
  sizesLoaded.value = true
  return sizes.value
}

async function createColor ({ name, hexValue }) {
  const color = await requestJSON(`${API_BASE}/colors`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, hex_value: hexValue }),
  })
  colors.value.push(color)
  return color
}

async function updateColor ({ id, name, hexValue }) {
  const color = await requestJSON(`${API_BASE}/colors`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, name, hex_value: hexValue }),
  })
  const idx = colors.value.findIndex(c => c.id === id)
  if (idx !== -1) {
    colors.value[idx] = color ?? { ...colors.value[idx], name, hex_value: hexValue }
  }
  return color
}

async function deleteColor (id) {
  await requestJSON(`${API_BASE}/colors/${encodeURIComponent(id)}`, { method: 'DELETE' })
  colors.value = colors.value.filter(c => c.id !== id)
}

async function createSize ({ name, type, order }) {
  const size = await requestJSON(`${API_BASE}/sizes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, type, order: String(order) }),
  })
  sizes.value.push(size)
  return size
}

async function updateSize ({ id, name, type, order }) {
  const size = await requestJSON(`${API_BASE}/sizes`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, name, type, order: String(order) }),
  })
  const idx = sizes.value.findIndex(s => s.id === id)
  if (idx !== -1) {
    sizes.value[idx] = size ?? { ...sizes.value[idx], name, type, order: String(order) }
  }
  return size
}

async function deleteSize (id) {
  await requestJSON(`${API_BASE}/sizes/${encodeURIComponent(id)}`, { method: 'DELETE' })
  sizes.value = sizes.value.filter(s => s.id !== id)
}

export function useColorsAndSizes () {
  return {
    colors,
    sizes,
    fetchColors,
    fetchSizes,
    createColor,
    updateColor,
    deleteColor,
    createSize,
    updateSize,
    deleteSize,
  }
}
