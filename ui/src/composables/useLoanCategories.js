import { createResource } from './useApiResource'

const resource = createResource({ path: '/loan_categories', listPath: '/loan_categories', rootKey: 'loan_categories' })

export function useLoanCategories () {
  return {
    loanCategories: resource.items,
    fetchLoanCategories: resource.fetchAll,
    createLoanCategory: resource.create,
    updateLoanCategory: resource.update,
    deleteLoanCategory: resource.remove,
  }
}
