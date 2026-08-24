import { API_BASE } from '@/config'
import { createResource } from './useApiResource'

const resource = createResource({ path: '/coupons', listPath: '/coupons', rootKey: 'coupons', idKey: 'code' })

async function deactivateCoupon (code) {
  await resource.requestJSON(`${API_BASE}/coupons/${encodeURIComponent(code)}/deactivate`, { method: 'PUT' })
  const idx = resource.items.value.findIndex(c => c.code === code)
  if (idx !== -1) {
    resource.items.value[idx] = { ...resource.items.value[idx], status: 'inactive' }
  }
}

export function useCoupons () {
  return {
    coupons: resource.items,
    fetchCoupons: resource.fetchAll,
    createCoupon: resource.create,
    deactivateCoupon,
  }
}
