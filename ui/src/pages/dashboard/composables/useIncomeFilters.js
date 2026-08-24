import { ref } from 'vue'

// Module-level (not per-component) so BranchMenu (the top filter bar) and
// any tab that reads it -- currently just Branches Daily Income -- share
// the same selection without prop drilling through dashboard.vue.
// null means "All Branches".
const selectedBranchId = ref(null)

export function useIncomeFilters () {
  return { selectedBranchId }
}
