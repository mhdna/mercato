import { API_BASE } from '@/config'
import { authFetch } from './useApi'

// Mirrors useBranches.js's requestJSON helper -- kept local rather than
// shared since it's a tiny, self-contained convention, not shared state.
async function requestJSON (url, options) {
  const res = await authFetch(url, options)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `Request failed with status ${res.status}`)
  }
  return res.json().catch(() => null)
}

export function useExpenseImages () {
  // Storage explorer feed -- see ListBranchExpenseImages's join in
  // db/query/branch_expense_image.sql, one round trip covers image +
  // parent expense + branch name for every row.
  async function fetchExpenseImages (pageId = 0, pageSize = 24) {
    const data = await requestJSON(`${API_BASE}/branch_expense_images?page_id=${pageId}&page_size=${pageSize}`)
    return { images: data?.images ?? [], total: data?.total ?? 0 }
  }

  // <img src> can't carry an Authorization header, so the file is fetched
  // as a blob and handed back as an object URL the caller can bind
  // directly -- same Bearer-token gate as every other admin endpoint,
  // rather than a separate unauthenticated static route.
  async function fetchExpenseImageObjectUrl (imageId) {
    const res = await authFetch(`${API_BASE}/branch_expense_images/${imageId}/file`)
    if (!res.ok) {
      throw new Error(`Failed to load image ${imageId}`)
    }
    const blob = await res.blob()
    return URL.createObjectURL(blob)
  }

  return { fetchExpenseImages, fetchExpenseImageObjectUrl }
}
