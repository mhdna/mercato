<template>
  <div class="activities-tab pa-4">
    <v-card class="activities-card" variant="outlined">
      <v-card-title class="d-flex align-center px-4 py-3">
        <v-icon class="me-2 text-medium-emphasis" icon="mdi-format-list-bulleted" size="20" />
        <div>
          <div class="text-subtitle-1 font-weight-medium">Recent activities</div>
          <div class="text-caption text-medium-emphasis">All activity recently synced from your branches</div>
        </div>
        <v-spacer />
        <v-btn
          aria-label="Refresh recent activities"
          icon="mdi-refresh"
          :loading="loading"
          size="small"
          variant="text"
          @click="load"
        />
      </v-card-title>

      <v-divider />

      <v-alert
        v-if="error"
        class="ma-3"
        density="compact"
        type="error"
        variant="tonal"
      >
        {{ error }}
      </v-alert>

      <v-data-table
        class="activities-table"
        density="comfortable"
        fixed-header
        :headers="headers"
        hover
        :items="activities"
        :items-per-page="14"
        :loading="loading"
        no-data-text="No synced activities yet"
      >
        <template #item.activity="{ item }">
          <div class="d-flex align-center ga-3">
            <v-avatar :color="item.color || 'info'" size="32" variant="tonal">
              <v-icon :icon="item.icon || 'mdi-information'" size="17" />
            </v-avatar>
            <span class="font-weight-medium">{{ item.activity || item.kind || 'Activity' }}</span>
          </div>
        </template>

        <template #item.branch_name="{ item }">
          <v-chip label size="small" variant="tonal">{{ item.branch_name || 'Unknown branch' }}</v-chip>
        </template>

        <template #item.details="{ item }">
          <span class="text-body-2">{{ item.details || item.label || '—' }}</span>
        </template>

        <template #item.amount="{ item }">
          <span v-if="hasAmount(item)" :class="amountClass(item.amount)">
            {{ money(item.amount, item.currency_code) }}
          </span>
          <span v-else class="text-medium-emphasis">—</span>
        </template>

        <template #item.synced_at="{ item }">
          <div class="text-body-2">{{ dateTime(item.synced_at || item.created_at || item.occurred_at) }}</div>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, onUnmounted, ref, watch } from 'vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { authFetch } from '@/composables/useApi'
  import { BRANCH_ACTIVITY_TYPES } from '@/composables/useBranchActivityMessage'
  import { API_BASE } from '@/config'
  import { useIncomeFilters } from '../composables/useIncomeFilters'

  const headers = [
    { title: 'Activity', key: 'activity', sortable: false },
    { title: 'Branch', key: 'branch_name', sortable: false },
    { title: 'Details', key: 'details', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Synced at', key: 'synced_at', align: 'end' },
  ]

  const activities = ref([])
  const loading = ref(false)
  const error = ref('')
  const { dashboardQuery } = useIncomeFilters()
  const { ensureConnected, onMessage } = useAdminSocket()

  let requestId = 0
  async function load () {
    const currentRequest = ++requestId
    loading.value = true
    error.value = ''
    try {
      const response = await authFetch(`${API_BASE}/dashboard/activities?${dashboardQuery()}&page_size=100&page_id=0`)
      if (!response.ok) {
        const body = await response.json().catch(() => ({}))
        throw new Error(body.error || `Failed to load recent activities (${response.status})`)
      }
      const data = await response.json()
      if (currentRequest === requestId) activities.value = data.activities ?? []
    } catch (error_) {
      if (currentRequest === requestId) error.value = error_.message
    } finally {
      if (currentRequest === requestId) loading.value = false
    }
  }

  let refreshTimer = null
  ensureConnected()
  const unsubscribe = onMessage(message => {
    if (!BRANCH_ACTIVITY_TYPES.includes(message?.type)) return
    clearTimeout(refreshTimer)
    refreshTimer = setTimeout(load, 500)
  })

  watch(dashboardQuery, load)
  onMounted(load)
  onUnmounted(() => {
    clearTimeout(refreshTimer)
    unsubscribe()
  })

  function money (amount, currency = '') {
    const value = Number(amount || 0) / 100
    const formatted = new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(Math.abs(value))
    return `${amount < 0 ? '−' : ''}${formatted} ${currency}`.trim()
  }

  function amountClass (amount) {
    if (amount > 0) return 'text-success font-weight-medium'
    if (amount < 0) return 'text-error font-weight-medium'
    return 'text-medium-emphasis'
  }

  function hasAmount (item) {
    return item.amount != null && !String(item.activity).startsWith('Attendance')
  }

  function dateTime (value) {
    return value ? new Date(value).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' }) : '—'
  }
</script>

<style scoped>
  .activities-tab {
    height: 100%;
    min-height: 0;
  }

  .activities-card {
    display: flex;
    height: 100%;
    min-height: 0;
    flex-direction: column;
  }

  .activities-table {
    min-height: 0;
  }
</style>
