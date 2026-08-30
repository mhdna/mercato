<template>
  <v-sheet class="dashboard d-flex flex-column" style="height: 100%">
    <div class="d-flex align-center flex-wrap justify-space-between">
      <div class="d-flex align-center flex-wrap">
        <v-btn-toggle
          v-model="scope"
          class="me-4 mt-2"
          density="compact"
          mandatory
          rounded="lg"
          variant="outlined"
        >
          <v-btn value="all">All</v-btn>
          <v-btn value="branch">Branches</v-btn>
          <v-btn value="admin">Admin</v-btn>
        </v-btn-toggle>
        <branch-menu />
      </div>
      <div class="d-flex align-center flex-wrap justify-end">
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

    <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
      <OverviewTab class="flex-grow-1" style="min-height: 0" />
    </div>
  </v-sheet>
</template>

<script setup>
  import BranchMenu from '@/components/Buttons/BranchMenu.vue'
  import { useIncomeFilters } from './composables/useIncomeFilters'
  import OverviewTab from './tabs/OverviewTab.vue'

  const { period, scope } = useIncomeFilters()
</script>
