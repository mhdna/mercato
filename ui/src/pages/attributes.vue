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
        <v-list
          v-model:selected="selection"
          class="attr-sidebar py-0"
          density="compact"
          mandatory
          nav
          select-strategy="single-leaf"
        >
          <v-list-item
            v-for="t in types"
            :key="t.id"
            :title="labelize(t.name)"
            :value="t.name"
          />
        </v-list>

        <v-divider vertical />

        <div class="attr-content">
          <div v-if="!selectedType" class="pa-6 text-medium-emphasis">
            Select an attribute on the left.
          </div>

          <v-data-table-server
            v-else
            v-model:items-per-page="itemsPerPage"
            v-model:page="page"
            v-model:sort-by="sortBy"
            class="attribute-table"
            density="comfortable"
            :headers="headers"
            :items="rows"
            :items-length="total"
            :items-per-page-options="[14, 25, 50, 100]"
            :loading="loading"
            @click:row="openRow"
            @update:options="loadRows"
          >
            <template #item.created_at="{ item }">
              {{ formatDate(item.created_at) }}
            </template>
          </v-data-table-server>
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
  import { useAttributes } from '@/composables/useAttributes'

  const {
    types,
    fetchTypes,
    fetchValuesPage,
    createValue,
    updateValue,
    deleteValue,
  } = useAttributes()

  const loading = ref(false)
  const loadError = ref('')
  const selection = ref([])
  const search = ref('')
  const rows = ref([])
  const total = ref(0)
  const page = ref(1)
  const itemsPerPage = ref(14)
  const sortBy = ref([{ key: 'created_at', order: 'desc' }])
  const lastOptions = ref(null)

  const selectedType = computed(() => selection.value[0] ?? null)

  async function loadRows (options) {
    lastOptions.value = options
    if (!selectedType.value) {
      rows.value = []
      total.value = 0
      return
    }

    loading.value = true
    loadError.value = ''
    try {
      const result = await fetchValuesPage({
        ...options,
        attribute: selectedType.value,
        search: search.value,
      })
      rows.value = result.items
      total.value = result.total
    } catch (error) {
      loadError.value = error.message
      rows.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  function reloadRows () {
    if (lastOptions.value) {
      loadRows({ ...lastOptions.value, page: page.value })
    }
  }

  function reloadFromFirstPage () {
    if (page.value !== 1) {
      page.value = 1
      return
    }
    reloadRows()
  }

  let searchTimer
  watch(search, () => {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(reloadFromFirstPage, 300)
  })

  watch(selectedType, () => {
    search.value = ''
    reloadFromFirstPage()
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
    loading.value = true
    try {
      await fetchTypes(true)
      if (selection.value.length === 0 && types.value.length > 0) {
        selection.value = [types.value[0].name]
      }
    } catch (error) {
      loadError.value = error.message
    } finally {
      loading.value = false
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

  function openRow (_event, { item }) {
    openEdit(item)
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
.attr-sidebar {
  width: 220px;
  flex: 0 0 220px;
  overflow-y: auto;
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
