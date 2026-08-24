import { createResource } from './useApiResource'

const resource = createResource({ path: '/cashbox_accounts', rootKey: 'cashbox_accounts' })

export function useCashboxAccounts () {
  return {
    cashboxAccounts: resource.items,
    fetchCashboxAccounts: resource.fetchAll,
    createCashboxAccount: resource.create,
    updateCashboxAccount: resource.update,
  }
}
