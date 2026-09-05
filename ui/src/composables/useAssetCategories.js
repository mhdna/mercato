import { createResource } from './useApiResource'

const resource = createResource({ path: '/asset_categories', listPath: '/asset_categories', rootKey: 'asset_categories' })

export function useAssetCategories () {
  return {
    assetCategories: resource.items,
    fetchAssetCategories: resource.fetchAll,
    createAssetCategory: resource.create,
    updateAssetCategory: resource.update,
    deleteAssetCategory: resource.remove,
  }
}
