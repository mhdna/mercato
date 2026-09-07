<template>
  <v-sheet class="dashboard d-flex flex-column" style="height: 100%">
    <div class="d-flex align-center flex-wrap justify-space-between">
      <v-tabs v-model="tab" color="primary">
        <v-tab class="px-8" value="overview">Overview</v-tab>
        <v-tab value="activities">Recent Activities</v-tab>
        <v-tab value="sales">Sales</v-tab>
        <v-tab value="financials">Financials</v-tab>
        <v-tab value="targets">Targets</v-tab>
        <v-tab value="visitors">Visitors</v-tab>
      </v-tabs>

      <div class="d-flex align-center flex-wrap justify-end">
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
    </div>

    <v-window v-model="tab" class="dashboard-window flex-grow-1" style="min-height: 0">
      <v-window-item class="h-100" value="overview">
        <div class="d-flex flex-column h-100" style="min-height: 0">
          <OverviewTab class="flex-grow-1" style="min-height: 0" />
        </div>
      </v-window-item>
      <v-window-item class="h-100" value="sales">
        <SalesTab class="h-100" />
      </v-window-item>
      <v-window-item class="h-100" value="activities">
        <RecentActivitiesTab class="h-100" />
      </v-window-item>
      <v-window-item class="h-100" value="financials" />
      <v-window-item class="h-100" value="targets">
        <TargetsTab class="h-100" />
      </v-window-item>
      <v-window-item class="h-100" value="visitors">
        <VisitorsTab class="h-100" />
      </v-window-item>
    </v-window>
  </v-sheet>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import BranchMenu from '@/components/Buttons/BranchMenu.vue'
  import { useIncomeFilters } from './composables/useIncomeFilters'
  import OverviewTab from './tabs/OverviewTab.vue'
  import RecentActivitiesTab from './tabs/RecentActivitiesTab.vue'
  import SalesTab from './tabs/SalesTab.vue'
  import TargetsTab from './tabs/TargetsTab.vue'
  import VisitorsTab from './tabs/VisitorsTab.vue'

  const tab = ref('overview')
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

  /* v-window inserts its own container between the dashboard and the active
     tab. Keep that container in the height chain so tab content can flex-fill
     the space left below the toolbar. */
  .dashboard-window :deep(.v-window__container) {
    height: 100%;
  }

</style>
