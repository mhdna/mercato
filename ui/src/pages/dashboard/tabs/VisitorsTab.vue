<template>
  <div class="d-flex flex-column flex-grow-1 pa-4" style="min-height: 0; overflow-y: auto">
    <v-alert v-if="error" class="mb-4" type="error" variant="tonal">{{ error }}</v-alert>

    <div v-if="loading" class="d-flex justify-center pa-8">
      <v-progress-circular color="primary" indeterminate />
    </div>

    <v-alert v-else-if="branchCards.length === 0" type="info" variant="tonal">
      No visitor counts reported yet. Branches send these from the nav-drawer people counter in the POS.
    </v-alert>

    <div v-else class="visitors-grid">
      <v-card v-for="card in branchCards" :key="card.branchId" class="pa-4">
        <v-card-title class="text-subtitle-1 px-0">{{ card.name }}</v-card-title>

        <div class="d-flex ga-6 mb-3">
          <div>
            <div class="text-caption text-medium-emphasis">In</div>
            <div class="text-h6">{{ card.totalIn }}</div>
          </div>
          <div>
            <div class="text-caption text-medium-emphasis">Out</div>
            <div class="text-h6">{{ card.totalOut }}</div>
          </div>
          <div>
            <div class="text-caption text-medium-emphasis">Net</div>
            <div class="text-h6" :class="card.totalIn - card.totalOut < 0 ? 'text-error' : ''">
              {{ card.totalIn - card.totalOut }}
            </div>
          </div>
        </div>

        <v-divider class="mb-2" />

        <v-table density="compact">
          <thead>
            <tr>
              <th class="text-left">Day</th>
              <th class="text-right">In</th>
              <th class="text-right">Out</th>
              <th class="text-right">Net</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in card.days" :key="row.day">
              <td>{{ row.day }}</td>
              <td class="text-right">{{ row.in_count }}</td>
              <td class="text-right">{{ row.out_count }}</td>
              <td class="text-right">{{ row.in_count - row.out_count }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-card>
    </div>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchVisitors } from '@/composables/useBranchVisitors'

  const { branches, fetchBranches } = useBranches()
  const { listVisitorDays } = useBranchVisitors()

  const loading = ref(false)
  const error = ref('')
  const days = ref([])

  onMounted(async () => {
    loading.value = true
    error.value = ''
    try {
      await fetchBranches()
      // One request: the newest 100 (branch, day) rollup rows across all
      // branches, grouped into per-branch cards below.
      days.value = await listVisitorDays()
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  })

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  const branchCards = computed(() => {
    const byBranch = new Map()
    for (const row of days.value) {
      if (!byBranch.has(row.branch_id)) byBranch.set(row.branch_id, [])
      byBranch.get(row.branch_id).push(row)
    }
    return [...byBranch.entries()]
      .map(([branchId, rows]) => ({
        branchId,
        name: branchName(branchId),
        days: rows,
        totalIn: rows.reduce((sum, r) => sum + r.in_count, 0),
        totalOut: rows.reduce((sum, r) => sum + r.out_count, 0),
      }))
      .toSorted((a, b) => a.name.localeCompare(b.name))
  })
</script>

<style scoped>
.visitors-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
}
</style>
