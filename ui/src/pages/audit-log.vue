<template>
  <v-dialog v-model="detailsDialog" max-width="640">
    <v-card v-if="detailsTarget">
      <v-card-title>{{ detailsTarget.entity_type }} #{{ detailsTarget.entity_id }} — {{ detailsTarget.action }}</v-card-title>
      <v-card-text>
        <v-row density="compact">
          <v-col cols="6"><span class="text-caption">Actor</span><div>{{ detailsTarget.actor_name || '—' }}</div></v-col>
          <v-col cols="6"><span class="text-caption">When</span><div>{{ new Date(detailsTarget.created_at).toLocaleString() }}</div></v-col>
        </v-row>
        <v-divider class="my-2" />
        <v-row>
          <v-col cols="6">
            <div class="text-caption mb-1">Before</div>
            <pre class="diff-block">{{ formatJSON(detailsTarget.before) }}</pre>
          </v-col>
          <v-col cols="6">
            <div class="text-caption mb-1">After</div>
            <pre class="diff-block">{{ formatJSON(detailsTarget.after) }}</pre>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="detailsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="audit-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-history" />
        <span>Audit Log</span>
        <v-spacer />
        <v-select
          v-model="filters.entityType"
          class="filter-select"
          clearable
          density="compact"
          hide-details
          :items="entityTypes"
          label="Entity"
          variant="outlined"
        />
        <v-select
          v-model="filters.action"
          class="filter-select"
          clearable
          density="compact"
          hide-details
          :items="actions"
          label="Action"
          variant="outlined"
        />
      </v-card-title>
      <v-divider />

      <div class="audit-content">
        <div class="pa-4">
          <ServerSideTable
            :api-u-r-l="apiURL"
            density="comfortable"
            flush
            :headers="headers"
            :query-params="{ entity_type: filters.entityType, action: filters.action }"
            root-key="audit_logs"
            :show-search-icon="false"
            @row-click="openDetails"
          >
            <template #item.action="{ item }">
              <v-chip :color="actionColor(item.action)" size="small">{{ item.action }}</v-chip>
            </template>
            <template #item.actor_name="{ item }">
              {{ item.actor_name || '—' }}
            </template>
            <template #item.created_at="{ item }">
              {{ new Date(item.created_at).toLocaleString() }}
            </template>
          </ServerSideTable>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { API_BASE } from '@/config'

  const apiURL = `${API_BASE}/audit_logs`
  const entityTypes = ['client', 'product', 'supplier', 'expense', 'price_list', 'branch_settings']
  const actions = ['create', 'update', 'delete']
  const filters = ref({ entityType: null, action: null })

  const headers = [
    { title: 'When', key: 'created_at', sortable: false },
    { title: 'Entity', key: 'entity_type', sortable: false },
    { title: 'ID', key: 'entity_id', align: 'end', sortable: false },
    { title: 'Action', key: 'action', sortable: false },
    { title: 'Actor', key: 'actor_name', sortable: false },
    { title: 'Source', key: 'source', sortable: false },
  ]

  function actionColor (a) {
    return { create: 'success', update: 'warning', delete: 'error' }[a] || 'grey'
  }

  const detailsDialog = ref(false)
  const detailsTarget = ref(null)

  function openDetails (item) {
    detailsTarget.value = item
    detailsDialog.value = true
  }

  function formatJSON (raw) {
    if (!raw) return '—'
    try {
      return JSON.stringify(typeof raw === 'string' ? JSON.parse(raw) : raw, null, 2)
    } catch {
      return String(raw)
    }
  }
</script>

<style scoped>
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.audit-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.audit-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.filter-select {
  flex: 0 0 200px;
}
.diff-block {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.75rem;
}
</style>
