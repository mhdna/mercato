import { ref } from 'vue'

export function useDashboardTabs () {
  const tab = ref('overview')

  const tabComponents = {
    overview: () => import('./tabs/OverviewTab.vue'),
    alerts: () => import('./tabs/AlertsTab.vue'),
    sales: () => import('./tabs/SalesAnalyticsTab.vue'),
    health: () => import('./tabs/HealthTab.vue'),
    demand: () => import('./tabs/DemandTab.vue'),
  }

  return { tab, tabComponents }
}
