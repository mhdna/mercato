import { createResource } from './useApiResource'

const resource = createResource({ path: '/cashboxes', rootKey: 'cashboxes' })

export function useCashboxes () {
  return {
    cashboxes: resource.items,
    fetchCashboxes: resource.fetchAll,
    createCashbox: resource.create,
    updateCashbox: resource.update,
  }
}
