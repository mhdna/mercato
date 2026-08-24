import { createResource } from './useApiResource'

const resource = createResource({ path: '/inventories', rootKey: 'inventories' })

export function useInventories () {
  return {
    inventories: resource.items,
    fetchInventories: resource.fetchAll,
    createInventory: resource.create,
    updateInventory: resource.update,
    deleteInventory: resource.remove,
  }
}
