<template>
  <div class="page-root">
    <v-card class="health-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-heart-pulse" />
        <span>Stock Health</span>
        <v-chip :color="statusColor" size="small" variant="flat">
          <v-icon icon="mdi-circle" size="10" start />
          {{ statusLabel }}
        </v-chip>
        <v-spacer />
        <v-btn icon="mdi-refresh" :loading="loading" variant="text" @click="load" />
      </v-card-title>
      <v-divider />

      <div class="health-content">
        <div class="pa-4">
          <div v-if="loading && !loaded" class="d-flex justify-center pa-12">
            <v-progress-circular color="primary" indeterminate />
          </div>
          <v-alert v-else-if="loadError" type="error" variant="tonal">{{ loadError }}</v-alert>

          <template v-else>
            <v-row class="mb-2" dense>
              <v-col cols="12" sm="4" md="3">
                <v-card class="summary-card gauge-card" variant="tonal">
                  <v-card-subtitle class="pt-3">Days of Inventory Left</v-card-subtitle>
                  <VChart autoresize class="gauge-chart" :option="gaugeOption" />
                </v-card>
              </v-col>
              <v-col cols="12" sm="8" md="9">
                <v-row dense>
                  <v-col cols="12" sm="4">
                    <v-card class="summary-card" variant="tonal">
                      <v-card-item>
                        <template #prepend>
                          <v-avatar color="secondary" variant="tonal">
                            <v-icon icon="mdi-cash-multiple" />
                          </v-avatar>
                        </template>
                        <v-card-title class="text-h4">{{ formatMoney(health.inventory_value) }}</v-card-title>
                        <v-card-subtitle>Inventory Value (cost)</v-card-subtitle>
                      </v-card-item>
                    </v-card>
                  </v-col>
                  <v-col cols="12" sm="4">
                    <v-card
                      class="summary-card kpi"
                      :class="{ 'kpi--active': slowMoverCategory === 'dead' }"
                      variant="tonal"
                      @click="toggleCategory('dead')"
                    >
                      <v-card-item>
                        <template #prepend>
                          <v-avatar color="error" variant="tonal">
                            <v-icon icon="mdi-skull-outline" />
                          </v-avatar>
                        </template>
                        <v-card-title class="text-h4">{{ health.dead_stock_count ?? '—' }}</v-card-title>
                        <v-card-subtitle>Dead Stock (0 sales / 90d)</v-card-subtitle>
                      </v-card-item>
                    </v-card>
                  </v-col>
                  <v-col cols="12" sm="4">
                    <v-card
                      class="summary-card kpi"
                      :class="{ 'kpi--active': slowMoverCategory === 'overstocked' }"
                      variant="tonal"
                      @click="toggleCategory('overstocked')"
                    >
                      <v-card-item>
                        <template #prepend>
                          <v-avatar color="warning" variant="tonal">
                            <v-icon icon="mdi-tray-full" />
                          </v-avatar>
                        </template>
                        <v-card-title class="text-h4">{{ health.overstocked_count ?? '—' }}</v-card-title>
                        <v-card-subtitle>Overstocked (90+ days on hand)</v-card-subtitle>
                      </v-card-item>
                    </v-card>
                  </v-col>
                </v-row>
              </v-col>
            </v-row>

            <v-alert v-if="statusNote" class="mb-2" :type="statusAlertType" variant="tonal">
              {{ statusNote }}
            </v-alert>

            <v-alert v-if="recommendation" class="mb-2" icon="mdi-lightbulb-outline" type="info" variant="tonal">
              {{ recommendation }}
            </v-alert>

            <div class="d-flex align-center ga-3 mt-4 mb-2">
              <span class="text-subtitle-1 font-weight-medium">Recommended For Sale</span>
              <v-chip-group v-model="categoryIndex" mandatory selected-class="text-primary">
                <v-chip filter size="small" variant="tonal">All</v-chip>
                <v-chip filter size="small" variant="tonal">Dead Stock</v-chip>
                <v-chip filter size="small" variant="tonal">Overstocked</v-chip>
              </v-chip-group>
              <v-spacer />
              <v-text-field
                v-model="slowMoverSearch"
                clearable
                density="compact"
                hide-details
                label="Search products"
                prepend-inner-icon="mdi-magnify"
                style="max-width: 280px"
                variant="outlined"
              />
            </div>
            <ServerSideTable
              :api-u-r-l="slowMoversURL"
              density="comfortable"
              :external-search="slowMoverSearch"
              flush
              :headers="slowMoverHeaders"
              :query-params="{ category: slowMoverCategory }"
              root-key="products"
              :show-search-icon="false"
            >
              <template #item.name="{ item }">
                <div>
                  <div>{{ item.name }}</div>
                  <div class="text-caption text-medium-emphasis">{{ item.code }}</div>
                </div>
              </template>
              <template #item.category="{ item }">
                <v-chip :color="item.category === 'dead' ? 'error' : 'warning'" size="small">
                  {{ item.category === 'dead' ? 'Never Sold (90d)' : 'Overstocked' }}
                </v-chip>
              </template>
              <template #item.days_of_inventory="{ item }">
                {{ item.days_of_inventory ?? 'Never sold' }}
              </template>
            </ServerSideTable>

            <div class="d-flex align-center ga-3 mt-6 mb-2">
              <span class="text-subtitle-1 font-weight-medium">By Location</span>
              <v-spacer />
              <v-text-field
                v-model="locationSearch"
                clearable
                density="compact"
                hide-details
                label="Search inventories"
                prepend-inner-icon="mdi-magnify"
                style="max-width: 280px"
                variant="outlined"
              />
            </div>
            <ServerSideTable
              :api-u-r-l="byInventoryURL"
              density="comfortable"
              :external-search="locationSearch"
              flush
              :headers="locationHeaders"
              root-key="inventories"
              :show-search-icon="false"
            >
              <template #item.type="{ item }">
                <span class="text-capitalize">{{ item.type }}</span>
              </template>
              <template #item.inventory_value="{ item }">
                {{ formatMoney(item.inventory_value) }}
              </template>
              <template #item.days_of_inventory="{ item }">
                <span :class="daysOfInventoryClass(item.days_of_inventory)">
                  {{ item.days_of_inventory ?? '—' }}
                </span>
              </template>
            </ServerSideTable>
          </template>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import VChart from 'vue-echarts'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useStockHealth } from '@/composables/useStockHealth'
  import { API_BASE } from '@/config'

  const { fetchStockHealth } = useStockHealth()

  const health = ref({})
  const loading = ref(false)
  const loaded = ref(false)
  const loadError = ref('')

  const byInventoryURL = `${API_BASE}/stock_health/by_inventory`
  const locationSearch = ref('')
  const locationHeaders = [
    { title: 'Inventory', key: 'name', align: 'start' },
    { title: 'Type', key: 'type', align: 'start' },
    { title: 'Units On Hand', key: 'total_units', align: 'end' },
    { title: 'Value (cost)', key: 'inventory_value', align: 'end' },
    { title: 'Units Sold (30d)', key: 'units_sold_30d', align: 'end' },
    { title: 'Days of Inventory', key: 'days_of_inventory', align: 'end' },
  ]

  const slowMoversURL = `${API_BASE}/stock_health/slow_movers`
  const slowMoverSearch = ref('')
  const slowMoverHeaders = [
    { title: 'Product', key: 'name', align: 'start' },
    { title: 'On Hand', key: 'on_hand', align: 'end' },
    { title: 'Sold (90d)', key: 'units_sold_90d', align: 'end' },
    { title: 'Days of Inventory', key: 'days_of_inventory', align: 'end' },
    { title: 'Why', key: 'category', align: 'start' },
  ]

  const categoryIndex = ref(0)
  const slowMoverCategory = computed(() => [null, 'dead', 'overstocked'][categoryIndex.value])
  function toggleCategory (category) {
    const idx = { dead: 1, overstocked: 2 }[category]
    categoryIndex.value = categoryIndex.value === idx ? 0 : idx
  }

  async function load () {
    loading.value = true
    loadError.value = ''
    try {
      health.value = await fetchStockHealth() ?? {}
      loaded.value = true
    } catch (error) {
      loadError.value = error.message
    } finally {
      loading.value = false
    }
  }

  const statusColor = computed(() => ({ healthy: 'success', warning: 'warning', critical: 'error' }[health.value.status] || 'grey'))
  const statusLabel = computed(() => ({ healthy: 'Healthy', warning: 'Needs Attention', critical: 'Critical' }[health.value.status] || '—'))
  const statusAlertType = computed(() => (health.value.status === 'critical' ? 'error' : 'warning'))
  const statusNote = computed(() => {
    if (health.value.status === 'critical') {
      return 'Inventory would sell through very soon at the current rate, or a large share of the catalog is dead stock.'
    }
    if (health.value.status === 'warning') {
      return 'Inventory would sell through within two weeks at the current rate, or a meaningful share of the catalog is slow-moving.'
    }
    return ''
  })

  // A short, concrete suggestion pulled from the same counts driving the
  // cards above -- "recommendation" the user asked this page to surface.
  const recommendation = computed(() => {
    const dead = health.value.dead_stock_count ?? 0
    const over = health.value.overstocked_count ?? 0
    if (dead === 0 && over === 0) return ''
    const parts = []
    if (dead > 0) parts.push(`${dead} product${dead === 1 ? '' : 's'} haven't sold at all in 90 days`)
    if (over > 0) parts.push(`${over} ${dead > 0 ? 'more are' : 'product' + (over === 1 ? ' is' : 's are')} carrying 90+ days of stock`)
    return `${parts.join(', and ')}. Consider a clearance sale or a purchasing pause on these -- see "Recommended For Sale" below.`
  })

  function daysOfInventoryClass (days) {
    if (days == null) return ''
    if (days < 3) return 'text-error font-weight-bold'
    if (days < 14) return 'text-warning font-weight-bold'
    return 'text-success'
  }

  function formatMoney (cents) {
    return (Number(cents ?? 0) / 100).toLocaleString(undefined, { style: 'currency', currency: 'USD' })
  }

  // Gauge zones mirror the status thresholds: red under 3 days, orange
  // under 14, green beyond. The axis caps at a soft max so the needle
  // stays legible even when on-hand would last months; "no sales at all"
  // (days_of_inventory is null) pins the needle at max with its own label.
  const gaugeMax = computed(() => {
    const days = health.value.days_of_inventory
    return Math.max(30, Math.ceil((days ?? 30) / 10) * 10 + 10)
  })
  const gaugeValue = computed(() => health.value.days_of_inventory ?? gaugeMax.value)
  const gaugeOption = computed(() => {
    const max = gaugeMax.value
    return {
      series: [
        {
          type: 'gauge',
          min: 0,
          max,
          radius: '90%',
          center: ['50%', '62%'],
          progress: { show: true, width: 10 },
          axisLine: {
            lineStyle: {
              width: 10,
              color: [
                [3 / max, '#EA4335'],
                [14 / max, '#FBBC05'],
                [1, '#34A853'],
              ],
            },
          },
          pointer: { show: true, length: '55%', width: 4 },
          axisTick: { show: false },
          splitLine: { show: false },
          axisLabel: { show: false },
          anchor: { show: true, size: 12, itemStyle: { color: '#888' } },
          title: { show: false },
          detail: {
            valueAnimation: true,
            fontSize: 22,
            fontWeight: 'bold',
            offsetCenter: [0, '75%'],
            formatter: () => (health.value.days_of_inventory == null ? 'No sales' : `${health.value.days_of_inventory}d`),
          },
          data: [{ value: gaugeValue.value }],
        },
      ],
    }
  })

  onMounted(load)
</script>

<style scoped>
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.health-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.health-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.summary-card :deep(.v-card-item) {
  padding: 12px 16px;
}
.gauge-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.gauge-chart {
  width: 100%;
  height: 140px;
}
.kpi {
  cursor: pointer;
  transition: outline-color 0.15s ease;
  outline: 2px solid transparent;
  outline-offset: -2px;
}
.kpi--active {
  outline-color: currentColor;
}
</style>
