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
  return res.json()
}

async function fetchColors (force = false) {
  if (colorsLoaded.value && !force) {
    return
  }
  const data = await requestJSON(`${API_BASE}/colors`)
  colors.value = data.colors ?? []
  colorsLoaded.value = true
}

async function fetchSizes (force = false) {
  if (sizesLoaded.value && !force) {
    return
  }
  const data = await requestJSON(`${API_BASE}/sizes`)
  sizes.value = data.sizes ?? []
  sizesLoaded.value = true
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

async function createSize ({ name, type, order }) {
  const size = await requestJSON(`${API_BASE}/sizes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, type, order }),
  })
  sizes.value.push(size)
  return size
}

export function useColorsAndSizes () {
  return { colors, sizes, fetchColors, fetchSizes, createColor, createSize }
}
