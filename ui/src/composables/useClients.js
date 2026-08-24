import { createResource } from './useApiResource'

const resource = createResource({ path: '/clients', rootKey: 'clients' })

export function useClients () {
  return {
    clients: resource.items,
    fetchClients: resource.fetchAll,
    createClient: resource.create,
    updateClient: resource.update,
    deleteClient: resource.remove,
  }
}
