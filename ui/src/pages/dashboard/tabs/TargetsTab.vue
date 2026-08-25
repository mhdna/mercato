<template>
  <div class="d-flex flex-column flex-grow-1 pa-4" style="min-height: 0; overflow-y: auto">
    <v-alert v-if="error" class="mb-4" type="error" variant="tonal">{{ error }}</v-alert>
    <div v-if="loading" class="d-flex justify-center pa-8">
      <v-progress-circular color="primary" indeterminate />
    </div>
    <v-alert v-else-if="branches.length === 0" type="info" variant="tonal">
      No branches yet.
    </v-alert>
    <div v-else class="targets-grid">
      <v-card v-for="branch in branches" :key="branch.id" class="pa-4">
        <v-card-title class="text-subtitle-1 px-0">{{ branch.name }}</v-card-title>
        <v-alert v-if="!(targetsByBranch[branch.id]?.length)" density="compact" type="info" variant="tonal">
          No targets set.
        </v-alert>
        <TargetProgressBars v-else :targets="targetsByBranch[branch.id]" />
      </v-card>
    </div>
  </div>
</template>

<script setup>
  import { onMounted, reactive, ref } from 'vue'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchTargets } from '@/composables/useBranchTargets'

  const { branches, fetchBranches } = useBranches()
  const { listBranchTargets } = useBranchTargets()

  const loading = ref(false)
  const error = ref('')
  const targetsByBranch = reactive({})

  onMounted(async () => {
    loading.value = true
    error.value = ''
    try {
      await fetchBranches()
      await Promise.all(branches.value.map(async branch => {
        targetsByBranch[branch.id] = await listBranchTargets(branch.id)
      }))
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  })
</script>

<style scoped>
.targets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
}
</style>
