<template>
  <div class="page-root">
    <v-alert
      v-if="loadError"
      class="mb-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ loadError }}
    </v-alert>

    <v-card class="attr-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-tag-multiple" />
        <span>Attributes</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="attribute-search"
          clearable
          density="compact"
          hide-details
          label="Search values"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn
          color="primary"
          :disabled="!selectedType"
          prepend-icon="mdi-plus"
          text="Add Value"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <div class="attr-layout">
        <PageSidebar v-model="selectedType" :items="typeItems" />

        <v-divider vertical />

        <div class="attr-content">
          <div v-if="!selectedType" class="pa-6 text-medium-emphasis">
            Select an attribute on the left.
          </div>

          <template v-else>
            <BulkDeleteBar
              :count="selectedRows.length"
              :error="selError"
              :loading="bulkDeleting"
              @clear="resetSel"
              @confirm="bulkDeleteRows"
            />
            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="`${API_BASE}/attributes/`"
              class="attribute-table"
              density="comfortable"
              :external-search="search ?? ''"
              flush
              :headers="headers"
              item-value="id"
              :query-params="{ attribute: selectedType }"
              root-key="values"
              selectable
              :show-search-icon="false"
              :sort-keys="['id', 'value', 'created_at']"
              @row-click="openEdit"
              @update:selected="v => (selectedRows = v)"
            >
              <template #item.created_at="{ item }">
                {{ formatDate(item.created_at) }}
              </template>
            </ServerSideTable>
          </template>
        </div>
      </div>
    </v-card>

    <v-dialog v-model="editDialog" max-width="420">
      <v-card>
        <v-card-title class="text-h6">
          {{ editing ? 'Edit Value' : `New ${labelize(selectedType)} Value` }}
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="draftValue"
            autofocus
            density="compact"
            hide-details="auto"
            label="Value"
            variant="outlined"
            @keyup.enter="saveValue"
          />
          <v-alert
            v-if="formError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ formError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-btn
            v-if="editing"
            color="error"
            prepend-icon="mdi-delete"
            text="Delete"
            variant="text"
            @click="deleteEditedValue"
          />
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="editDialog = false" />
          <v-btn
            color="primary"
            :loading="saving"
            text="Save"
            variant="flat"
            @click="saveValue"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="text-h6">Delete value</v-card-title>
        <v-card-text>
          Delete <strong>{{ deleteTarget?.value }}</strong>? This can't be undone.
          <v-alert
            v-if="deleteError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ deleteError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="deleteDialog = false" />
          <v-btn
            color="error"
            :loading="deleting"
            text="Delete"
            variant="flat"
            @click="doDelete"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import PageSidebar from '@/components/PageSidebar.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useAttributes } from '@/composables/useAttributes'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { API_BASE } from '@/config'

  const {
    types,
    fetchTypes,
    createValue,
    updateValue,
    deleteValue,
  } = useAttributes()

  const {
    selected: selectedRows,
    deleting: bulkDeleting,
    error: selError,
    reset: resetSel,
    run: runBulk,
  } = useBulkDelete()

  const tableRef = ref(null)
  function reloadRows () {
    tableRef.value?.reload()
  }

  async function bulkDeleteRows () {
    try {
      await runBulk('/attributes/bulk_delete', { ids: selectedRows.value })
      resetSel()
      reloadRows()
    } catch { /* error shown in the bar */ }
  }

  const loadError = ref('')
  const selectedType = ref(null)
  const search = ref('')

  const typeItems = computed(() => types.value.map(t => ({
    value: t.name,
    title: labelize(t.name),
  })))

  watch(selectedType, () => {
    search.value = ''
    resetSel()
  })

  const headers = [
    { title: 'Value', key: 'value', align: 'start' },
    { title: 'ID', key: 'id', align: 'start', width: 90 },
    { title: 'Created At', key: 'created_at', align: 'start', width: 200 },
  ]

  function labelize (name) {
    if (!name) return ''
    return name.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
  }

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
  }

  onMounted(async () => {
    try {
      await fetchTypes(true)
      if (selectedType.value == null && types.value.length > 0) {
        selectedType.value = types.value[0].name
      }
    } catch (error) {
      loadError.value = error.message
    }
  })

  // ---- create / edit -------------------------------------------------------
  const editDialog = ref(false)
  const editing = ref(null)
  const draftValue = ref('')
  const saving = ref(false)
  const formError = ref('')

  function openCreate () {
    editing.value = null
    draftValue.value = ''
    formError.value = ''
    editDialog.value = true
  }

  function openEdit (item) {
    editing.value = item
    draftValue.value = item.value
    formError.value = ''
    editDialog.value = true
  }

  async function saveValue () {
    const value = draftValue.value.trim()
    if (!value) {
      formError.value = 'Value is required.'
      return
    }
    saving.value = true
    formError.value = ''
    try {
      await (editing.value ? updateValue({ id: editing.value.id, value }) : createValue({ attribute: selectedType.value, value }))
      editDialog.value = false
      reloadRows()
    } catch (error) {
      formError.value = error.message
    } finally {
      saving.value = false
    }
  }

  // ---- delete ------------------------------------------------------------
  const deleteDialog = ref(false)
  const deleteTarget = ref(null)
  const deleting = ref(false)
  const deleteError = ref('')

  function confirmDelete (item) {
    deleteTarget.value = item
    deleteError.value = ''
    deleteDialog.value = true
  }

  function deleteEditedValue () {
    if (!editing.value) return
    editDialog.value = false
    confirmDelete(editing.value)
  }

  async function doDelete () {
    deleting.value = true
    deleteError.value = ''
    try {
      await deleteValue(deleteTarget.value.id)
      deleteDialog.value = false
      reloadRows()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }
</script>

<style scoped>
/* Fill the layout's scroll wrapper and keep the scroll inside the value list,
   so the sidebar and page header stay put. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.attr-card {
  flex: 1 1 auto;
  min-height: 0;
}
.attr-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.attr-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.attribute-search {
  flex: 0 1 320px;
}
:deep(.attribute-table tbody tr) {
  cursor: pointer;
}
</style>
