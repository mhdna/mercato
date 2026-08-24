import { createResource } from './useApiResource'

const resource = createResource({ path: '/invoice_types', listPath: '/invoice_types', rootKey: 'invoice_types' })

export function useInvoiceTypes () {
  return {
    invoiceTypes: resource.items,
    fetchInvoiceTypes: resource.fetchAll,
    createInvoiceType: resource.create,
    updateInvoiceType: resource.update,
  }
}
