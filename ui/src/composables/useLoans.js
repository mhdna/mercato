import { createResource } from './useApiResource'

// listPath asks for a large single page -- this composable backs the loan
// picker on the loan-payments page, not a paginated table (loans.vue's
// tables talk to /loans directly through ServerSideTable instead).
const resource = createResource({ path: '/loans', listPath: '/loans?page_size=100', rootKey: 'loans' })

export function useLoans () {
  return {
    loans: resource.items,
    fetchLoans: resource.fetchAll,
    createLoan: resource.create,
  }
}
