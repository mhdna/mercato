<template>
  <v-sheet class="d-flex flex-column fill-height">
    <v-tabs v-model="tab" color="primary">
      <v-tab value="sales">Sales Analytics</v-tab>
      <v-tab value="demand">Demand</v-tab>
    </v-tabs>

    <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
      <component :is="tabComponents[tab]" :key="tab" class="flex-grow-1" />
    </div>
  </v-sheet>
</template>

<script setup>
  import { defineAsyncComponent, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'

  const tabComponents = {
    sales: defineAsyncComponent(() => import('./dashboard/tabs/SalesAnalyticsTab.vue')),
    demand: defineAsyncComponent(() => import('./dashboard/tabs/DemandTab.vue')),
  }

  const route = useRoute()
  const router = useRouter()
  const routeTab = typeof route.query.tab === 'string' ? route.query.tab : null
  const tab = ref(
    routeTab && Object.prototype.hasOwnProperty.call(tabComponents, routeTab)
      ? routeTab
      : 'sales',
  )

  watch(tab, value => {
    if (route.query.tab !== value) {
      router.replace({ query: { ...route.query, tab: value } })
    }
  })

  watch(
    () => route.query.tab,
    value => {
      const nextTab = typeof value === 'string' && Object.prototype.hasOwnProperty.call(tabComponents, value)
        ? value
        : 'sales'

      if (tab.value !== nextTab) {
        tab.value = nextTab
      }
    },
  )
</script>
