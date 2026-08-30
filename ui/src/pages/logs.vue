<template>
  <div class="logs-page pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center justify-space-between ga-3 mb-5">
      <div>
        <h1 class="text-h5 font-weight-bold">Application logs</h1>
        <p class="text-body-2 text-medium-emphasis mt-1">
          Persisted backend activity and recent browser events, loaded in bounded batches.
        </p>
      </div>
      <div class="d-flex ga-2">
        <v-btn
          :loading="loadingServerLogs"
          prepend-icon="mdi-refresh"
          text="Refresh"
          variant="tonal"
          @click="loadServerLogs"
        />
        <v-btn
          :disabled="clientEntries.length === 0"
          prepend-icon="mdi-delete"
          text="Clear browser logs"
          variant="text"
          @click="clear"
        />
      </div>
    </div>

    <div class="summary-grid mb-5">
      <v-card v-for="item in summary" :key="item.label" class="summary-card" variant="outlined">
        <v-card-text>
          <div class="text-caption text-medium-emphasis">{{ item.label }}</div>
          <div :class="['text-h5 font-weight-bold mt-1', item.color && `text-${item.color}`]">
            {{ item.value }}
          </div>
        </v-card-text>
      </v-card>
    </div>

    <v-card class="mb-5" variant="outlined">
      <v-card-title class="text-subtitle-1">Session information</v-card-title>
      <v-divider />
      <v-card-text class="info-grid">
        <div v-for="item in sessionInfo" :key="item.label" class="info-item">
          <span class="text-caption text-medium-emphasis">{{ item.label }}</span>
          <span class="text-body-2 text-truncate" :title="item.value">{{ item.value }}</span>
        </div>
      </v-card-text>
    </v-card>

    <v-card class="log-card" variant="outlined">
      <v-alert
        v-if="serverError"
        class="ma-3 mb-0"
        density="compact"
        type="error"
        variant="tonal"
      >
        {{ serverError }}
      </v-alert>
      <div class="filters pa-3">
        <v-text-field
          v-model.trim="query"
          clearable
          density="compact"
          hide-details
          label="Search messages"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-select
          v-model="level"
          density="compact"
          hide-details
          :items="levelOptions"
          label="Level"
          variant="outlined"
        />
        <v-select
          v-model="source"
          density="compact"
          hide-details
          :items="sourceOptions"
          label="Source"
          variant="outlined"
        />
      </div>
      <v-divider />

      <div v-if="filteredEntries.length > 0" aria-live="polite" class="log-list selectable">
        <article
          v-for="entry in filteredEntries"
          :key="entry.id"
          :class="['log-entry', `log-entry--${entry.level}`]"
        >
          <time class="log-time" :datetime="entry.timestamp">{{ formatTime(entry.timestamp) }}</time>
          <v-chip :color="levelColor(entry.level)" label size="x-small" variant="tonal">
            {{ entry.level.toUpperCase() }}
          </v-chip>
          <span class="log-source">{{ entry.source }}</span>
          <pre class="log-message">{{ entry.message }}</pre>
        </article>
      </div>

      <div v-else class="empty-state text-center text-medium-emphasis pa-10">
        <v-icon class="mb-3" icon="mdi-text" size="40" />
        <div class="text-subtitle-1">{{ entries.length > 0 ? 'No matching logs' : 'No logs yet' }}</div>
        <div class="text-body-2 mt-1">
          {{ entries.length > 0 ? 'Try changing the search or level filter.' : 'New activity will appear here during this session.' }}
        </div>
      </div>

      <v-divider />
      <div class="d-flex justify-space-between text-caption text-medium-emphasis pa-3">
        <span>Showing {{ filteredEntries.length }} of {{ entries.length }}</span>
        <span>Server: last 200 · Browser: max {{ capacity }}</span>
      </div>
    </v-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue'
  import { authFetch } from '@/composables/useApi'
  import { useClientLogs } from '@/composables/useClientLogs'
  import { API_BASE } from '@/config'

  const { capacity, clear, entries: clientEntries } = useClientLogs()
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

  const entries = computed(() => [...clientEntries.value, ...serverEntries.value])

  const filteredEntries = computed(() => {
    const needle = query.value.toLowerCase()
    return entries.value.filter(entry => {
      const matchesLevel = level.value === 'all' || entry.level === level.value
      const matchesSource = source.value === 'all' || entry.group === source.value
      const matchesQuery = !needle || `${entry.message} ${entry.source}`.toLowerCase().includes(needle)
      return matchesLevel && matchesSource && matchesQuery
    }).toReversed()
  })

  const summary = computed(() => [
    { label: 'Loaded', value: entries.value.length },
    { label: 'Errors', value: entries.value.filter(item => item.level === 'error').length, color: 'error' },
    { label: 'Warnings', value: entries.value.filter(item => item.level === 'warn').length, color: 'warning' },
    { label: 'Backend entries', value: serverEntries.value.length },
  ])

  const sessionInfo = [
    { label: 'Environment', value: import.meta.env.MODE },
    { label: 'API server', value: API_BASE },
    { label: 'Page', value: window.location.origin },
    { label: 'Language', value: navigator.language },
    { label: 'Online', value: navigator.onLine ? 'Yes' : 'No' },
    { label: 'Platform', value: navigator.userAgentData?.platform || navigator.platform || 'Unknown' },
  ]

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
      serverEntries.value = (Array.isArray(data.logs) ? data.logs : []).slice(-200).map((line, index) => parseServerLine(line, index))
    } catch (error) {
      serverError.value = error.message || 'Could not load backend logs.'
    } finally {
      loadingServerLogs.value = false
    }
  }

  onMounted(loadServerLogs)
</script>

<style scoped>
.logs-page { max-width: 1500px; width: 100%; margin: 0 auto; }
.summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
.summary-card { border-top-width: 3px; }
.info-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 18px 28px; }
.info-item { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
.filters { display: grid; grid-template-columns: minmax(220px, 1fr) 180px 180px; gap: 12px; }
.log-list { max-height: min(58vh, 720px); overflow: auto; background: rgb(var(--v-theme-surface)); }
.log-entry { display: grid; grid-template-columns: 112px 72px 80px minmax(0, 1fr); align-items: start; gap: 10px; padding: 9px 14px; border-inline-start: 3px solid transparent; border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.log-entry--error { border-inline-start-color: rgb(var(--v-theme-error)); background: rgba(var(--v-theme-error), .06); }
.log-entry--warn { border-inline-start-color: rgb(var(--v-theme-warning)); background: rgba(var(--v-theme-warning), .06); }
.log-entry--info { border-inline-start-color: rgb(var(--v-theme-info)); }
.log-time, .log-source, .log-message { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 12px; }
.log-time, .log-source { padding-top: 2px; color: rgba(var(--v-theme-on-surface), .65); }
.log-source { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.log-message { margin: 0; overflow-wrap: anywhere; white-space: pre-wrap; }

@media (max-width: 800px) {
  .summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .info-grid { grid-template-columns: 1fr 1fr; }
  .filters { grid-template-columns: 1fr; }
  .log-entry { grid-template-columns: 90px 68px minmax(0, 1fr); }
  .log-source { display: none; }
}

@media (max-width: 500px) {
  .info-grid { grid-template-columns: 1fr; }
  .log-entry { grid-template-columns: 1fr auto; }
  .log-message { grid-column: 1 / -1; }
}
</style>
