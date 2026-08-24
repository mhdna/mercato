import { createResource } from './useApiResource'

const resource = createResource({ path: '/currencies', rootKey: 'currencies', idKey: 'code' })

export function useCurrencies () {
  return {
    currencies: resource.items,
    fetchCurrencies: resource.fetchAll,
    createCurrency: resource.create,
    updateCurrency: resource.update,
    deleteCurrency: resource.remove,
  }
}
