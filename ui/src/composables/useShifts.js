import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/shifts', listPath: '/shifts', rootKey: 'shifts' })

async function closeShift (id) {
  const item = await resource.requestJSON(`${API_BASE}/shifts/${id}/close`, { method: 'POST' })
  const idx = resource.items.value.findIndex(s => s.id === id)
  if (idx !== -1) {
    resource.items.value[idx] = { ...resource.items.value[idx], is_closed: true }
  }
  return item
}

export function useShifts () {
  return {
    shifts: resource.items,
    fetchShifts: resource.fetchAll,
    createShift: resource.create,
    closeShift,
  }
}
