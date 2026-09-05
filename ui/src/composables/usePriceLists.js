import { useValueListStore } from './useValueListStore'

// Price lists: line-item value is `price`, stored in cents.
const store = useValueListStore({
  base: '/price_lists',
  rootKey: 'price_lists',
  listIdKey: 'price_list_id',
  valueKey: 'price',
})

export function usePriceLists () {
  return {
    ...store,
    // Back-compat aliases for callers that predate useValueListStore
    // (sales-invoices.vue).
    priceLists: store.list,
    fetchPriceLists: store.fetchLists,
  }
}
