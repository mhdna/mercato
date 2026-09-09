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
              <v-col cols="12" sm="3">
                <v-card class="summary-card" variant="tonal">
                  <v-card-item>
                    <template #prepend>
                      <v-avatar color="primary" variant="tonal">
                        <v-icon icon="mdi-calendar-clock-outline" />
                      </v-avatar>
                    </template>
                    <v-card-title class="text-h4">{{ daysLabel }}</v-card-title>
                    <v-card-subtitle>Days of Inventory Left</v-card-subtitle>
                  </v-card-item>
                </v-card>
              </v-col>
              <v-col cols="12" sm="3">
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
              <v-col cols="12" sm="3">
                <v-card class="summary-card" variant="tonal">
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
              <v-col cols="12" sm="3">
                <v-card class="summary-card" variant="tonal">
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

            <v-alert v-if="statusNote" class="mb-2" :type="statusAlertType" variant="tonal">
              {{ statusNote }}
            </v-alert>

            <v-alert v-if="recommendation" class="mb-2" icon="mdi-lightbulb-outline" type="info" variant="tonal">
              {{ recommendation }}
            </v-alert>

            <div class="text-subtitle-1 font-weight-medium mt-4 mb-2">Recommended For Sale</div>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-card variant="outlined">
                  <v-card-subtitle class="d-flex align-center ga-2 pt-3">
                    <v-icon color="error" icon="mdi-skull-outline" size="18" />
                    Dead Stock -- No Sales in 90 Days
                  </v-card-subtitle>
                  <v-divider class="mt-2" />
                  <v-virtual-scroll
                    v-if="deadStock.length"
                    class="slow-mover-scroll"
                    :item-height="64"
                    :items="deadStock"
                  >
                    <template #default="{ item }">
                      <v-list-item class="px-4 py-2">
                        <template #prepend>
                          <v-avatar class="me-3" color="error" rounded="lg" size="40" variant="tonal">
                            <v-icon color="error" icon="mdi-skull-outline" size="20" />
                          </v-avatar>
                        </template>
                        <v-list-item-title class="font-weight-medium">{{ item.name }}</v-list-item-title>
                        <v-list-item-subtitle>{{ item.code }} · On hand: {{ item.on_hand }}</v-list-item-subtitle>
                        <template #append>
                          <span class="text-caption text-medium-emphasis">Never sold</span>
                        </template>
                      </v-list-item>
                      <v-divider />
                    </template>
                  </v-virtual-scroll>
                  <div v-else class="text-medium-emphasis text-center pa-6">No dead stock -- everything active is selling.</div>
                </v-card>
              </v-col>
              <v-col cols="12" md="6">
                <v-card variant="outlined">
                  <v-card-subtitle class="d-flex align-center ga-2 pt-3">
                    <v-icon color="warning" icon="mdi-tray-full" size="18" />
                    Overstocked -- 90+ Days On Hand
                  </v-card-subtitle>
                  <v-divider class="mt-2" />
                  <v-virtual-scroll
                    v-if="overstocked.length"
                    class="slow-mover-scroll"
                    :item-height="64"
                    :items="overstocked"
                  >
                    <template #default="{ item }">
                      <v-list-item class="px-4 py-2">
                        <template #prepend>
                          <v-avatar class="me-3" color="warning" rounded="lg" size="40" variant="tonal">
                            <v-icon color="warning" icon="mdi-tray-full" size="20" />
                          </v-avatar>
                        </template>
                        <v-list-item-title class="font-weight-medium">{{ item.name }}</v-list-item-title>
                        <v-list-item-subtitle>{{ item.code }} · On hand: {{ item.on_hand }} · Sold 90d: {{ item.units_sold_90d }}</v-list-item-subtitle>
                        <template #append>
                          <span class="text-caption text-medium-emphasis">{{ item.days_of_inventory }}d supply</span>
                        </template>
                      </v-list-item>
                      <v-divider />
                    </template>
                  </v-virtual-scroll>
                  <div v-else class="text-medium-emphasis text-center pa-6">Nothing overstocked right now.</div>
                </v-card>
              </v-col>
            </v-row>
          </template>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import { useStockHealth } from '@/composables/useStockHealth'
  import { authFetch } from '@/composables/useApi'
  import { API_BASE } from '@/config'

  const { fetchStockHealth } = useStockHealth()

  const health = ref({})
  const loading = ref(false)
  const loaded = ref(false)
  const loadError = ref('')

  // The two slow-mover columns: fetched once as a capped list and rendered
  // with v-virtual-scroll, so a large catalog stays smooth without paging
  // controls -- this is a "everything at a glance" view, not a browsable table.
  const SLOW_MOVER_CAP = 300
  const deadStock = ref([])
  const overstocked = ref([])

  async function fetchSlowMovers (category) {
    const params = new URLSearchParams({ category, page_size: String(SLOW_MOVER_CAP), page_id: '0' })
    const res = await authFetch(`${API_BASE}/stock_health/slow_movers?${params}`)
    if (!res.ok) throw new Error(`Request failed with status ${res.status}`)
    const data = await res.json().catch(() => null)
    return data?.products ?? []
  }

  async function load () {
    loading.value = true
    loadError.value = ''
    try {
      const [healthData, dead, over] = await Promise.all([
        fetchStockHealth(),
        fetchSlowMovers('dead'),
        fetchSlowMovers('overstocked'),
      ])
      health.value = healthData ?? {}
      deadStock.value = dead
      overstocked.value = over
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

  const daysLabel = computed(() => health.value.days_of_inventory ?? '—')

  function formatMoney (cents) {
    return (Number(cents ?? 0) / 100).toLocaleString(undefined, { style: 'currency', currency: 'USD' })
  }

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
.slow-mover-scroll {
  height: 420px;
}
</style>
