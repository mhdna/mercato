import { createResource } from './useApiResource'

const resource = createResource({
  path: '/coupon_categories',
  listPath: '/coupon_categories',
  rootKey: 'coupon_categories',
})

export function useCouponCategories () {
  return {
    couponCategories: resource.items,
    fetchCouponCategories: resource.fetchAll,
    createCouponCategory: resource.create,
    updateCouponCategory: resource.update,
    deleteCouponCategory: resource.remove,
  }
}
