<template>
  <div class="page-root">
    <v-card class="alerts-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-alert-circle-outline" />
        <span>Alerts</span>
        <v-chip-group v-model="viewIndex" class="ms-2" mandatory selected-class="text-primary">
          <v-chip filter size="small" variant="tonal">Active</v-chip>
          <v-chip filter size="small" variant="tonal">
            Ignored
            <v-badge v-if="ignoredCount" class="ml-2" color="grey" :content="ignoredCount" inline />
          </v-chip>
        </v-chip-group>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="alerts-search"
          clearable
          density="compact"
          hide-details
          label="Search products"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-text-field
          v-model.number="defaultThresholdInput"
          density="compact"
          hide-details
          label="Default threshold"
          style="max-width: 180px"
          type="number"
          variant="outlined"
          @keyup.enter="saveDefaultThreshold"
        >
          <template #append-inner>
            <v-icon-btn
              v-if="defaultThresholdDirty"
              icon="mdi-check"
              :loading="savingDefault"
              size="small"
              variant="text"
              @click="saveDefaultThreshold"
            />
          </template>
        </v-text-field>
      </v-card-title>
      <v-divider />

      <div class="alerts-content">
        <div class="pa-4">
          <!-- Click a card to filter the table below to that bucket -->
          <v-row v-if="!showIgnored" class="mb-2" dense>
            <v-col v-for="kpi in kpis" :key="kpi.bucket" cols="12" sm="4">
              <v-card
                class="summary-card kpi"
                :class="{ 'kpi--active': bucket === kpi.bucket }"
                :color="kpi.color"
                variant="tonal"
                @click="toggleBucket(kpi.bucket)"
              >
                <v-card-item>
                  <template #prepend>
                    <v-avatar :color="kpi.color" variant="tonal">
                      <v-icon :icon="kpi.icon" />
                    </v-avatar>
                  </template>
                  <v-card-title class="text-h4">{{ kpi.value ?? '—' }}</v-card-title>
                  <v-card-subtitle>{{ kpi.label }}</v-card-subtitle>
                </v-card-item>
              </v-card>
            </v-col>
          </v-row>

          <v-alert v-if="showIgnored" class="mb-2" density="compact" type="info" variant="tonal">
            These products are excluded from low-stock alerts. Unmute one to start tracking it again.
          </v-alert>

          <ServerSideTable
            :key="showIgnored"
            ref="tableRef"
            :api-u-r-l="apiURL"
            density="comfortable"
            :external-search="search"
            flush
            :headers="headers"
            hover
            :query-params="showIgnored ? {} : { status: bucket === 'all' ? null : bucket }"
            root-key="products"
            :show-search-icon="false"
          >
            <template #item.name="{ item }">
              <div class="d-flex align-center ga-2">
                <v-icon
                  v-if="!showIgnored"
                  :color="item.on_hand <= 0 ? 'error' : 'warning'"
                  :icon="item.on_hand <= 0 ? 'mdi-close-circle' : 'mdi-alert'"
                  size="small"
                />
                <v-icon v-else color="grey" icon="mdi-bell-off-outline" size="small" />
                <div>
                  <div>{{ item.name }}</div>
                  <div class="text-caption text-medium-emphasis">{{ item.code }}</div>
                </div>
              </div>
            </template>
            <template #item.on_hand="{ item }">
              <span :class="item.on_hand <= 0 ? 'text-error font-weight-bold' : 'text-warning font-weight-bold'">
                {{ item.on_hand }}
              </span>
            </template>
            <template #item.threshold="{ item }">
              <div class="d-flex align-center ga-1">
                <v-text-field
                  :model-value="displayThreshold(item)"
                  density="compact"
                  hide-details
                  reverse
                  style="max-width: 90px"
                  type="number"
                  @keyup.enter="saveThreshold(item)"
                  @update:model-value="v => (thresholdInputs[item.id] = v)"
                />
                <v-icon-btn
                  v-if="thresholdDirty(item)"
                  icon="mdi-check"
                  :loading="savingId === item.id"
                  size="small"
                  variant="text"
                  @click="saveThreshold(item)"
                />
                <v-chip v-if="!hasOverride(item)" class="ml-1" size="x-small" variant="tonal">default</v-chip>
                <v-icon-btn
                  v-else
                  icon="mdi-restore"
                  size="small"
                  title="Reset to default"
                  variant="text"
                  @click="clearThreshold(item)"
                />
              </div>
            </template>
            <template #item.status="{ item }">
              <v-chip :color="item.on_hand <= 0 ? 'error' : 'warning'" size="small">
                {{ item.on_hand <= 0 ? 'Out of Stock' : 'Low Stock' }}
              </v-chip>
            </template>
            <template #item.mute="{ item }">
              <v-icon-btn
                v-if="!showIgnored"
                icon="mdi-bell-off-outline"
                :loading="mutingId === item.id"
                size="small"
                title="Don't alert on this product"
                variant="text"
                @click="muteProduct(item)"
              />
              <v-icon-btn
                v-else
                icon="mdi-bell-outline"
                :loading="mutingId === item.id"
                size="small"
                title="Resume alerting on this product"
                variant="text"
                @click="unmuteProduct(item)"
              />
            </template>
          </ServerSideTable>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useAppSettings } from '@/composables/useAppSettings'
  import { useLowStockAlerts } from '@/composables/useLowStockAlerts'
  import { API_BASE } from '@/config'

  const { getAppSettings } = useAppSettings()
  const { fetchSummary, fetchIgnoredCount, setProductThreshold, setDefaultThreshold, setAlertsEnabled } = useLowStockAlerts()

  const search = ref('')
  const tableRef = ref(null)

  const viewIndex = ref(0)
  const showIgnored = computed(() => viewIndex.value === 1)
  const apiURL = computed(() => `${API_BASE}${showIgnored.value ? '/low_stock/ignored' : '/low_stock'}`)

  const headers = computed(() => showIgnored.value
    ? [
      { title: 'Product', key: 'name', sortable: false },
      { title: 'On Hand', key: 'on_hand', align: 'end', sortable: false },
      { title: '', key: 'mute', align: 'end', sortable: false, width: 56 },
    ]
    : [
      { title: 'Product', key: 'name', sortable: false },
      { title: 'On Hand', key: 'on_hand', align: 'end', sortable: false },
      { title: 'Threshold', key: 'threshold', align: 'end', sortable: false, width: 220 },
      { title: 'Status', key: 'status', sortable: false },
      { title: '', key: 'mute', align: 'end', sortable: false, width: 56 },
    ])

  const summary = ref({})
  const ignoredCount = ref(0)
  async function loadSummary () {
    try {
      summary.value = await fetchSummary() ?? {}
    } catch {
      summary.value = {}
    }
  }

  const bucket = ref('all')
  function toggleBucket (b) {
    bucket.value = bucket.value === b ? 'all' : b
  }

  const kpis = computed(() => [
    {
      bucket: 'out_of_stock',
      label: 'Out of Stock',
      icon: 'mdi-close-circle-outline',
      color: 'error',
      value: summary.value.out_of_stock_count,
    },
    {
      bucket: 'low_stock',
      label: 'Low Stock',
      icon: 'mdi-alert-outline',
      color: 'warning',
      value: summary.value.low_stock_count,
    },
    {
      bucket: 'all',
      label: 'Active Products Monitored',
      icon: 'mdi-package-variant',
      color: 'primary',
      value: summary.value.total_products,
    },
  ])

  const defaultThresholdInput = ref(null)
  const defaultThresholdSaved = ref(null)
  const defaultThresholdDirty = computed(() =>
    defaultThresholdInput.value != null
    && Number(defaultThresholdInput.value) !== Number(defaultThresholdSaved.value),
  )
  const savingDefault = ref(false)

  async function loadDefaultThreshold () {
    try {
      const settings = await getAppSettings()
      defaultThresholdSaved.value = settings?.default_low_stock_threshold ?? 5
      defaultThresholdInput.value = defaultThresholdSaved.value
    } catch {
      defaultThresholdSaved.value = 5
      defaultThresholdInput.value = 5
    }
  }

  async function saveDefaultThreshold () {
    if (defaultThresholdInput.value == null) return
    savingDefault.value = true
    try {
      const settings = await setDefaultThreshold(Number(defaultThresholdInput.value))
      defaultThresholdSaved.value = settings?.default_low_stock_threshold ?? defaultThresholdInput.value
      tableRef.value?.reload()
      loadSummary()
    } finally {
      savingDefault.value = false
    }
  }

  // Go's sql.NullInt64 marshals as {Int64, Valid} (no json tags on that
  // type) -- mirrors the nullableString/displayUnitCost pattern used
  // elsewhere in this app for the same reason.
  function nullableInt (value) {
    if (value == null) return null
    if (typeof value !== 'object') return value
    const valid = value.Valid ?? value.valid
    if (valid === false) return null
    return value.Int64 ?? value.int64 ?? null
  }

  function hasOverride (item) {
    return nullableInt(item.low_stock_threshold) != null
  }

  function displayThreshold (item) {
    if (thresholdInputs[item.id] != null) return thresholdInputs[item.id]
    return nullableInt(item.low_stock_threshold) ?? item.effective_threshold
  }

  const thresholdInputs = reactive({})
  const savingId = ref(null)

  function thresholdDirty (item) {
    const current = nullableInt(item.low_stock_threshold) ?? item.effective_threshold
    return thresholdInputs[item.id] != null && Number(thresholdInputs[item.id]) !== Number(current)
  }

  async function saveThreshold (item) {
    const value = thresholdInputs[item.id]
    if (value == null || value === '') return
    savingId.value = item.id
    try {
      await setProductThreshold(item.id, Number(value))
      delete thresholdInputs[item.id]
      tableRef.value?.reload()
      loadSummary()
    } finally {
      savingId.value = null
    }
  }

  async function clearThreshold (item) {
    savingId.value = item.id
    try {
      await setProductThreshold(item.id, null)
      delete thresholdInputs[item.id]
      tableRef.value?.reload()
      loadSummary()
    } finally {
      savingId.value = null
    }
  }

  const mutingId = ref(null)

  async function loadIgnoredCount () {
    try {
      ignoredCount.value = await fetchIgnoredCount()
    } catch {
      ignoredCount.value = 0
    }
  }

  async function muteProduct (item) {
    mutingId.value = item.id
    try {
      await setAlertsEnabled(item.id, false)
      tableRef.value?.reload()
      loadSummary()
      loadIgnoredCount()
    } finally {
      mutingId.value = null
    }
  }

  async function unmuteProduct (item) {
    mutingId.value = item.id
    try {
      await setAlertsEnabled(item.id, true)
      tableRef.value?.reload()
      loadSummary()
      loadIgnoredCount()
    } finally {
      mutingId.value = null
    }
  }

  onMounted(() => {
    loadSummary()
    loadDefaultThreshold()
    loadIgnoredCount()
  })
</script>

<style scoped>
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.alerts-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.alerts-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.alerts-search {
  flex: 0 1 280px;
}
.summary-card :deep(.v-card-item) {
  padding: 12px 16px;
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
