import { createResource } from './useApiResource'

// Users are POS operators identified by name + a 4-6 digit PIN. The list route
// is registered without a trailing slash (see api/server.go), hence listPath.
const resource = createResource({ path: '/users', listPath: '/users', rootKey: 'users' })

export function useUsers () {
  return {
    users: resource.items,
    fetchUsers: resource.fetchAll,
    createUser: resource.create,
    updateUser: resource.update,
    deleteUser: resource.remove,
  }
}
