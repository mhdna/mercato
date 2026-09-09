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
                        <v-icon icon="mdi-close-circle-outline" />
                      </v-avatar>
                    </template>
                    <v-card-title class="text-h4">{{ health.out_of_stock_count ?? '—' }}</v-card-title>
                    <v-card-subtitle>Out of Stock</v-card-subtitle>
                  </v-card-item>
                </v-card>
              </v-col>
              <v-col cols="12" sm="3">
                <v-card class="summary-card" variant="tonal">
                  <v-card-item>
                    <template #prepend>
                      <v-avatar color="warning" variant="tonal">
                        <v-icon icon="mdi-alert-outline" />
                      </v-avatar>
                    </template>
                    <v-card-title class="text-h4">{{ health.low_stock_count ?? '—' }}</v-card-title>
                    <v-card-subtitle>Low Stock</v-card-subtitle>
                  </v-card-item>
                </v-card>
              </v-col>
            </v-row>

            <v-alert v-if="statusNote" class="mb-2" :type="statusAlertType" variant="tonal">
              {{ statusNote }}
            </v-alert>

            <div class="d-flex align-center ga-3 mt-4 mb-2">
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
      return 'Inventory is critically low: several products are out of stock, or on-hand would sell through in under 3 days at the current rate.'
    }
    if (health.value.status === 'warning') {
      return 'Some products are low or out of stock, or inventory would sell through within two weeks at the current rate.'
    }
    return ''
  })

  const daysLabel = computed(() => health.value.days_of_inventory ?? '—')

  function daysOfInventoryClass (days) {
    if (days == null) return ''
    if (days < 3) return 'text-error font-weight-bold'
    if (days < 14) return 'text-warning font-weight-bold'
    return 'text-success'
  }

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
</style>
