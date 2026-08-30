<template>
  <div class="text-center">
    <v-menu>
      <template #activator="{ props }">
        <v-btn
          class="me-4 mt-2"
          rounded="lg"
          v-bind="props"
          style="border-color: rgb(var(--v-theme-surface-light)); width: 120px;"
          variant="outlined"
        >
          {{ selectedLabel }}
        </v-btn>
      </template>

      <v-list>
        <v-list-item
          :active="selectedBranchId === null"
          @click="selectedBranchId = null"
        >
          <v-list-item-title>All Branches</v-list-item-title>
        </v-list-item>
        <v-list-item
          v-for="branch in branches"
          :key="branch.id"
          :active="selectedBranchId === branch.id"
          @click="selectedBranchId = branch.id"
        >
          <v-list-item-title>{{ branch.name }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script setup>
  import { computed, onMounted } from 'vue'
  import { useBranches } from '@/composables/useBranches'
  import { useIncomeFilters } from '@/pages/dashboard/composables/useIncomeFilters'

  const { branches, fetchBranches } = useBranches()
  const { selectedBranchId } = useIncomeFilters()

  onMounted(() => fetchBranches())

  const selectedLabel = computed(() => {
    if (selectedBranchId.value === null) return 'All Branches'
    return branches.value.find(b => b.id === selectedBranchId.value)?.name ?? 'All Branches'
  })
</script>
