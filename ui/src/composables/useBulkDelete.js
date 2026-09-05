import { ref } from 'vue'
import { API_BASE } from '@/config'
import { authFetch } from './useApi'

/**
 * Shared state + helper for a table's checkbox multi-select + bulk delete.
 * `selected` is bound to a v-data-table's v-model (an array of item-value
 * keys -- ids, codes, product_ids, ...). `run(path, body)` POSTs to a
 * `/x/bulk_delete` endpoint and throws on failure with the server's
 * `error` message (so a 409 "still in use" surfaces verbatim).
 */
export function useBulkDelete () {
  const selected = ref([])
  const deleting = ref(false)
  const error = ref('')

  function reset () {
    selected.value = []
    error.value = ''
  }

  async function run (path, body) {
    deleting.value = true
    error.value = ''
    try {
      const res = await authFetch(`${API_BASE}${path}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      return res.json().catch(() => null)
    } catch (error_) {
      error.value = error_.message
      throw error_
    } finally {
      deleting.value = false
    }
  }

  return { selected, deleting, error, reset, run }
}
