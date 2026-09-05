import { createResource } from './useApiResource'

// listPath asks for the full currency list (not a 10-row page) -- this
// composable backs the CurrencySelect picker; currencies.vue's table talks
// to /currencies/ through ServerSideTable for its own pagination.
const resource = createResource({ path: '/currencies', listPath: '/currencies/?all=true', rootKey: 'currencies', idKey: 'code' })

export function useCurrencies () {
  return {
    currencies: resource.items,
    fetchCurrencies: resource.fetchAll,
    createCurrency: resource.create,
    updateCurrency: resource.update,
    deleteCurrency: resource.remove,
  }
}
