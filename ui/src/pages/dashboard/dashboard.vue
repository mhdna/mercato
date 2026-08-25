<template>
  <v-sheet class="dashboard d-flex flex-column" style="height: 100%">
    <div class="d-flex justify-space-between">
      <v-tabs v-model="tab" color="primary">
        <v-tab class="px-8" value="overview">Overview</v-tab>
        <v-tab class="d-flex align-center justify-space-between" value="income">
          <div>Daily Income</div>
        </v-tab>
        <v-tab
          class="d-flex align-center justify-space-between px-6 pe-8"
          value="alerts"
        >
          <div>Alerts</div>
          <v-badge class="mx-4 me-0 mb-1" color="red-darken-4" content="118" />
        </v-tab>
        <v-tab value="sales">Sales Analytics</v-tab>
        <v-tab
          class="d-flex align-center justify-space-between me-4"
          value="health"
        >
          <div class="me-4">Health</div>
          <v-badge class="mb-1" color="green-darken-2" />
        </v-tab>
        <v-tab value="demand">Demand</v-tab>
        <v-tab value="targets">Targets</v-tab>
      </v-tabs>
      <!-- <v-btn
        rounded="lg"
        variant="outlined"
        class="me-4 mt-2"
        style="border-color: rgb(var(--v-theme-surface-light));"
      >
      Filtered by: Last 30 days
      </v-btn> -->
      <!-- <filter-btn/> -->
      <!--TODO: only filters for overview tab-->
      <div class="d-flex">
        <branch-menu />
        <v-btn-toggle
          v-model="period"
          class="me-4 mt-2"
          density="compact"
          mandatory
          rounded="lg"
          variant="outlined"
        >
          <!-- <v-btn value="this-month">Monthly</v-btn> -->
          <v-btn value="yesterday">Yesterday</v-btn>
          <v-btn value="today">Today</v-btn>
          <v-btn value="30d">30d</v-btn>
        </v-btn-toggle>
      </div>
    </div>

    <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
      <component
        :is="tabComponents[tab]"
        :key="tab"
        class="flex-grow-1"
        style="min-height: 0"
      />
    </div>
  </v-sheet>
</template>

<script setup>
  import { defineAsyncComponent, ref } from 'vue'
  import BranchMenu from '@/components/Buttons/BranchMenu.vue'
  import FilterBtn from '../FilterBtn.vue'

  const tab = ref('overview')
  const period = ref('today')

  const tabComponents = {
    overview: defineAsyncComponent(() => import('./tabs/OverviewTab.vue')),
    alerts: defineAsyncComponent(() => import('./tabs/AlertsTab.vue')),
    income: defineAsyncComponent(() => import('./tabs/IncomeTab.vue')),
    sales: defineAsyncComponent(() => import('./tabs/SalesAnalyticsTab.vue')),
    health: defineAsyncComponent(() => import('./tabs/HealthTab.vue')),
    demand: defineAsyncComponent(() => import('./tabs/DemandTab.vue')),
    targets: defineAsyncComponent(() => import('./tabs/TargetsTab.vue')),
  }
</script>
