<template>
  <div class="page-root">
    <v-card class="logs-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-text" />
        <span>Application logs</span>
        <v-chip
          v-if="errorCount"
          color="error"
          label
          size="small"
          variant="tonal"
        >
          {{ errorCount }} errors
        </v-chip>
        <v-chip
          v-if="warnCount"
          color="warning"
          label
          size="small"
          variant="tonal"
        >
          {{ warnCount }} warnings
        </v-chip>
        <v-spacer />
        <v-text-field
          v-model.trim="query"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search messages"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-select
          v-model="level"
          class="filter-select"
          density="compact"
          hide-details
          :items="levelOptions"
          label="Level"
          variant="outlined"
        />
        <v-select
          v-model="source"
          class="filter-select"
          density="compact"
          hide-details
          :items="sourceOptions"
          label="Source"
          variant="outlined"
        />
        <v-btn
          icon="mdi-refresh"
          :loading="loadingServerLogs"
          size="small"
          title="Refresh backend logs"
          variant="text"
          @click="loadServerLogs"
        />
        <v-btn
          :disabled="clientEntries.length === 0"
          icon="mdi-delete"
          size="small"
          title="Clear browser logs"
          variant="text"
          @click="clear"
        />
      </v-card-title>
      <v-divider />

      <v-alert
        v-if="serverError"
        class="ma-3 mb-0"
        density="compact"
        type="error"
        variant="tonal"
      >
        {{ serverError }}
      </v-alert>

      <div class="logs-content">
        <v-data-table
          class="logs-table"
          density="comfortable"
          :headers="headers"
          :items="filteredEntries"
          :items-per-page="25"
          :sort-by="[{ key: 'timestamp', order: 'desc' }]"
        >
          <template #item.timestamp="{ item }">
            <span class="mono">{{ formatTime(item.timestamp) }}</span>
          </template>
          <template #item.level="{ item }">
            <v-chip :color="levelColor(item.level)" label size="x-small" variant="tonal">
              {{ item.level.toUpperCase() }}
            </v-chip>
          </template>
          <template #item.source="{ item }">
            <span class="mono text-medium-emphasis">{{ item.source }}</span>
          </template>
          <template #item.message="{ item }">
            <pre class="log-message">{{ item.message }}</pre>
          </template>
          <template #no-data>
            <div class="empty-state text-center text-medium-emphasis pa-10">
              <v-icon class="mb-3" icon="mdi-text" size="40" />
              <div class="text-subtitle-1">
                {{ entries.length > 0 ? 'No matching logs' : 'No logs yet' }}
              </div>
              <div class="text-body-2 mt-1">
                {{ entries.length > 0
                  ? 'Try changing the search or filters.'
                  : 'New activity will appear here during this session.' }}
              </div>
            </div>
          </template>
        </v-data-table>
      </div>
    </v-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue'
  import { authFetch } from '@/composables/useApi'
  import { useClientLogs } from '@/composables/useClientLogs'
  import { API_BASE } from '@/config'

  const { clear, entries: clientEntries } = useClientLogs()
  const serverEntries = ref([])
  const loadingServerLogs = ref(false)
  const serverError = ref('')
  const query = ref('')
  const level = ref('all')
  const source = ref('all')

  const levelOptions = [
    { title: 'All levels', value: 'all' },
    { title: 'Errors', value: 'error' },
    { title: 'Warnings', value: 'warn' },
    { title: 'Info', value: 'info' },
  ]
  const sourceOptions = [
    { title: 'All sources', value: 'all' },
    { title: 'Go backend', value: 'backend' },
    { title: 'Browser', value: 'browser' },
  ]

  const headers = [
    { title: 'Time', key: 'timestamp', width: 130 },
    { title: 'Level', key: 'level', width: 90 },
    { title: 'Source', key: 'source', width: 120 },
    { title: 'Message', key: 'message', sortable: false },
  ]

  const entries = computed(() => [...clientEntries.value, ...serverEntries.value])

  const filteredEntries = computed(() => {
    const needle = query.value.toLowerCase()
    return entries.value.filter(entry => {
      const matchesLevel = level.value === 'all' || entry.level === level.value
      const matchesSource = source.value === 'all' || entry.group === source.value
      const matchesQuery = !needle
        || `${entry.message} ${entry.source}`.toLowerCase().includes(needle)
      return matchesLevel && matchesSource && matchesQuery
    })
  })

  const errorCount = computed(() => entries.value.filter(item => item.level === 'error').length)
  const warnCount = computed(() => entries.value.filter(item => item.level === 'warn').length)

  function formatTime (timestamp) {
    if (!timestamp) return '—'
    try {
      return new Intl.DateTimeFormat(undefined, {
        hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3,
      }).format(new Date(timestamp))
    } catch {
      return '—'
    }
  }

  function levelColor (value) {
    return { error: 'error', warn: 'warning', info: 'info' }[value] || 'default'
  }

  function parseServerLine (line, index) {
    const timestampMatch = line.match(/^(\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2})/)
    const normalized = line.toLowerCase()
    const level = normalized.includes('error') || normalized.includes('panic') || normalized.includes('[recovery]')
      ? 'error'
      : (normalized.includes('warn') ? 'warn' : 'info')
    return {
      id: `server-${index}-${line.slice(0, 24)}`,
      group: 'backend',
      level,
      message: line,
      source: 'backend',
      timestamp: timestampMatch ? timestampMatch[1].replaceAll('/', '-').replace(' ', 'T') : '',
    }
  }

  async function loadServerLogs () {
    loadingServerLogs.value = true
    serverError.value = ''
    try {
      const response = await authFetch(`${API_BASE}/logs?limit=200`)
      if (!response.ok) throw new Error(`Server returned ${response.status}`)
      const data = await response.json()
      serverEntries.value = (Array.isArray(data.logs) ? data.logs : [])
        .slice(-200)
        .map((line, index) => parseServerLine(line, index))
    } catch (error) {
      serverError.value = error.message || 'Could not load backend logs.'
    } finally {
      loadingServerLogs.value = false
    }
  }

  onMounted(loadServerLogs)
</script>

<style scoped>
/* Fill the layout's flex-column scroll wrapper so the scroll lives inside the
   table, not the whole page. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.logs-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.logs-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.table-search {
  flex: 0 1 280px;
}
.filter-select {
  flex: 0 0 150px;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  white-space: nowrap;
}
.log-message {
  margin: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>
