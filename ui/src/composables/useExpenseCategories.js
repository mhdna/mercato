import { createResource } from './useApiResource'

const resource = createResource({ path: '/expense_categories', listPath: '/expense_categories', rootKey: 'expense_categories' })

export function useExpenseCategories () {
  return {
    expenseCategories: resource.items,
    fetchExpenseCategories: resource.fetchAll,
    createExpenseCategory: resource.create,
    updateExpenseCategory: resource.update,
    deleteExpenseCategory: resource.remove,
  }
}
