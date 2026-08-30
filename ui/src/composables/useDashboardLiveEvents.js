import { onUnmounted } from 'vue'
import { useIncomeFilters } from '@/pages/dashboard/composables/useIncomeFilters'
import { useAdminSocket } from './useAdminSocket'

// Admin WS pushes that change what the Overview tab shows -- new branch
// sales/returns, branch expenses, branch loans, or a receipt photo landing
// on an existing expense. (Central admin invoices/expenses are entered in
// this same app, so their own pages already refresh; there's no push for
// them.)
const RELEVANT = new Set([
  'branch_invoice_created',
  'branch_expense_created',
  'branch_loan_created',
  'branch_expense_images_uploaded',
])

// Calls `onRefresh` (trailing-debounced) whenever a relevant admin push
// arrives for the branch currently in view. A branch catch-up can fire a
// burst of these, so the debounce collapses them into one reload.
export function useDashboardLiveEvents (onRefresh, { debounceMs = 600 } = {}) {
  const { selectedBranchId } = useIncomeFilters()
  const { ensureConnected, onMessage } = useAdminSocket()

  ensureConnected()

  let timer = null
  const unsubscribe = onMessage(message => {
    if (!RELEVANT.has(message?.type)) {
      return
    }
    if (
      selectedBranchId.value != null
      && message.branch_id != null
      && message.branch_id !== selectedBranchId.value
    ) {
      return
    }

    clearTimeout(timer)
    timer = setTimeout(onRefresh, debounceMs)
  })

  onUnmounted(() => {
    clearTimeout(timer)
    unsubscribe()
  })
}
