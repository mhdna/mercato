import { createResource } from './useApiResource'

const resource = createResource({ path: '/sales_invoices', listPath: '/sales_invoices', rootKey: 'sales_invoices' })

export function useSalesInvoices () {
  return {
    salesInvoices: resource.items,
    fetchSalesInvoices: resource.fetchAll,
    createSalesInvoice: resource.create,
  }
}
