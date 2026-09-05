import { useValueListStore } from './useValueListStore'

// Discount lists: line-item value is `discount`, stored as a whole percent.
const store = useValueListStore({
  base: '/discount_lists',
  rootKey: 'discount_lists',
  listIdKey: 'discount_list_id',
  valueKey: 'discount',
})

export function useDiscountLists () {
  return {
    ...store,
    discountLists: store.list,
    fetchDiscountLists: store.fetchLists,
  }
}
