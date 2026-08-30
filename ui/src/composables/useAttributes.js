import { computed, ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Attribute *types* are a fixed set seeded server-side (category, brand, …).
// Attribute *values* are the free-form options that belong to each type and
// that products get tagged with.
const types = ref([])
const values = ref([])
const typesLoaded = ref(false)
const valuesLoaded = ref(false)

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

async function fetchTypes (force = false) {
  if (typesLoaded.value && !force) {
    return types.value
  }
  const data = await requestJSON(`${API_BASE}/attribute_types`)
  types.value = data?.attributes ?? []
  typesLoaded.value = true
  return types.value
}

async function fetchValues (force = false) {
  if (valuesLoaded.value && !force) {
    return values.value
  }
  const data = await requestJSON(`${API_BASE}/attribute_values`)
  values.value = data?.values ?? []
  valuesLoaded.value = true
  return values.value
}

async function fetchAll (force = false) {
  await Promise.all([fetchTypes(force), fetchValues(force)])
}

async function fetchValuesPage ({ attribute, page, itemsPerPage, search = '', sortBy = [] }) {
  const url = new URL(`${API_BASE}/attributes/`)
  url.searchParams.set('attribute', attribute)
  url.searchParams.set('page_size', itemsPerPage)
  url.searchParams.set('page_id', (page - 1) * itemsPerPage)
  if (search.trim()) {
    url.searchParams.set('search', search.trim())
  }
  if (sortBy.length > 0) {
    url.searchParams.set('sort_by', sortBy[0].key)
    url.searchParams.set('sort_order', sortBy[0].order)
  }

  const data = await requestJSON(url.toString())
  return {
    items: data?.values ?? [],
    total: data?.total ?? 0,
  }
}

// `attribute` is the type *name* (the API resolves it to an id and upserts).
async function createValue ({ attribute, value }) {
  const created = await requestJSON(`${API_BASE}/attributes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ attribute, value }),
  })
  const row = { ...created, attribute_name: attribute }
  const idx = values.value.findIndex(v => v.id === row.id)
  if (idx === -1) {
    values.value.push(row)
  } else {
    values.value[idx] = row
  }
  return row
}

async function updateValue ({ id, value }) {
  const updated = await requestJSON(`${API_BASE}/attributes`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, value }),
  })
  const idx = values.value.findIndex(v => v.id === id)
  if (idx !== -1) {
    values.value[idx] = { ...values.value[idx], ...updated }
  }
  return updated
}

async function deleteValue (id) {
  await requestJSON(`${API_BASE}/attributes/${encodeURIComponent(id)}`, { method: 'DELETE' })
  values.value = values.value.filter(v => v.id !== id)
}

// { [attributeName]: AttributeValue[] } — every known type gets a key even if empty.
const valuesByType = computed(() => {
  const grouped = {}
  for (const t of types.value) {
    grouped[t.name] = []
  }
  for (const v of values.value) {
    (grouped[v.attribute_name] ||= []).push(v)
  }
  for (const key of Object.keys(grouped)) {
    grouped[key].sort((a, b) => a.value.localeCompare(b.value))
  }
  return grouped
})

export function useAttributes () {
  return {
    types,
    values,
    valuesByType,
    fetchTypes,
    fetchValues,
    fetchValuesPage,
    fetchAll,
    createValue,
    updateValue,
    deleteValue,
  }
}
