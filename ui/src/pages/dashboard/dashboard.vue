<template>
  <v-sheet class="dashboard d-flex flex-column" style="height: 100%">
    <div class="d-flex align-center flex-wrap justify-end pt-2">
      <v-menu>
        <template #activator="{ props }">
          <v-btn
            class="me-4 mt-2"
            rounded="lg"
            v-bind="props"
            style="border-color: rgb(var(--v-theme-surface-light)); width: 120px;"
            variant="outlined"
          >
            {{ scopeLabel }}
          </v-btn>
        </template>

        <v-list>
          <v-list-item
            v-for="item in scopeItems"
            :key="item.value"
            :active="scope === item.value"
            @click="scope = item.value"
          >
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
      <branch-menu />
      <v-btn-toggle
        v-model="period"
        class="me-4 mt-2"
        density="compact"
        mandatory
        rounded="lg"
        variant="outlined"
      >
        <v-btn value="this-month">This month</v-btn>
        <v-btn value="30d">30D</v-btn>
        <v-btn value="last-month">Last month</v-btn>
        <v-btn value="this-year">This year</v-btn>
      </v-btn-toggle>
    </div>

    <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
      <OverviewTab class="flex-grow-1" style="min-height: 0" />
    </div>
  </v-sheet>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import BranchMenu from '@/components/Buttons/BranchMenu.vue'
  import { useIncomeFilters } from './composables/useIncomeFilters'
  import OverviewTab from './tabs/OverviewTab.vue'

  const scopeItems = [
    { title: 'All Branches', value: 'all' },
    { title: 'Branches', value: 'branch' },
    { title: 'Admin', value: 'admin' },
  ]
  const { period, scope } = useIncomeFilters()
  const scopeLabel = computed(() => scopeItems.find(item => item.value === scope.value)?.title ?? 'All Branches')
</script>

<style scoped>
  /* Keep dashboard card outlines consistent with the filter controls. */
  .dashboard :deep(.v-card--variant-outlined),
  .dashboard :deep(.v-btn--variant-outlined),
  .dashboard :deep(.v-btn-toggle--variant-outlined),
  .dashboard :deep(.fin-position) {
    border-color: rgb(var(--v-theme-surface-light));
  }

  .dashboard :deep(.summary-card-body) {
    background-color: rgb(var(--v-theme-surface-light));
  }
</style>
