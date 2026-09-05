<template>
  <v-dialog v-model="dialog" max-width="560">
    <v-card class="px-4">
      <v-card-title>{{ editingId ? 'Edit Inventory' : 'Add a New Inventory' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
          />
          <v-select
            v-model="type.value.value"
            density="compact"
            :disabled="!!editingId"
            :error-messages="type.errorMessage.value"
            :items="['warehouse', 'store']"
            label="Type"
          />
          <v-text-field
            v-model="code.value.value"
            density="compact"
            :disabled="!!editingId"
            :error-messages="code.errorMessage.value"
            label="Code"
          />
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model.number="latitude.value.value"
                clearable
                density="compact"
                :disabled="!!editingId"
                label="Latitude (optional)"
                type="number"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="longitude.value.value"
                clearable
                density="compact"
                :disabled="!!editingId"
                label="Longitude (optional)"
                type="number"
              />
            </v-col>
          </v-row>

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-btn
              v-if="editingId"
              color="error"
              prepend-icon="mdi-delete"
              text="Delete"
              variant="text"
              @click="deleteFromEdit"
            />
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingId ? 'Save' : 'Add Inventory'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="deleteDialog" max-width="420">
    <v-card>
      <v-card-title>Delete Inventory</v-card-title>
      <v-card-text>Are you sure you want to delete "{{ deleteTarget?.name }}"? This cannot be undone.</v-card-text>
      <v-alert v-if="deleteError" class="mx-4 mb-2" type="error" variant="tonal">{{ deleteError }}</v-alert>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" @click="deleteDialog = false" />
        <v-btn color="error" :loading="deleting" text="Delete" @click="confirmDelete" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="inventory-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-warehouse" />
        <span>Inventory</span>
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Inventory"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <div class="inventory-content">
        <div class="pa-4">
          <BulkDeleteBar
            :count="selectedRows.length"
            :error="selError"
            :loading="bulkDeleting"
            @clear="clearSelection"
            @confirm="bulkDeleteRows"
          />

          <ServerSideTable
            ref="tableRef"
            :api-u-r-l="apiURL"
            class="inventory-table"
            density="comfortable"
            flush
            :headers="headers"
            :max-page-size="10"
            root-key="inventories"
            selectable
            :show-search-icon="false"
            @row-click="openEdit"
            @update:selected="selectedRows = $event"
          >
            <template #item.actions="{ item }">
              <v-icon-btn icon="mdi-warehouse" size="small" variant="text" @click.stop="openStock(item)" />
            </template>
          </ServerSideTable>
        </div>
      </div>
    </v-card>
  </div>

  <v-dialog v-model="stockDialog" max-width="820">
    <v-card class="px-4">
      <v-card-title>Stock — {{ stockTarget?.name }}</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="stockSearch"
          class="mb-2"
          clearable
          density="compact"
          hide-details
          placeholder="Search SKU / barcode"
          @update:model-value="onStockSearch"
        />
        <v-alert v-if="stockError" class="mb-2" type="error" variant="tonal">{{ stockError }}</v-alert>
        <v-table density="compact">
          <thead>
            <tr>
              <th>SKU</th>
              <th>Barcode</th>
              <th class="text-end">On hand</th>
              <th class="text-end">Avg cost</th>
              <th class="text-end" style="width: 220px;">Adjust</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in stockRows" :key="row.variant_id">
              <td>
                {{ row.product_code }} — {{ row.product_name }}
                <span v-if="row.color_name || row.size_name" class="text-medium-emphasis">
                  ({{ [row.color_name, row.size_name].filter(Boolean).join(' / ') }})
                </span>
              </td>
              <td>{{ row.barcode }}</td>
              <td class="text-end" :class="row.quantity < 0 ? 'text-error' : ''">{{ row.quantity }}</td>
              <td class="text-end">{{ row.avg_cost }}</td>
              <td class="text-end">
                <v-text-field
                  v-model.number="adjustInputs[row.variant_id]"
                  density="compact"
                  hide-details
                  placeholder="set count"
                  style="max-width: 120px; display: inline-block;"
                  type="number"
                />
                <v-btn
                  class="ml-1"
                  :loading="adjustingId === row.variant_id"
                  size="small"
                  text="Set"
                  @click="applyCount(row)"
                />
              </td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="stockDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useInventories } from '@/composables/useInventories'
  import { API_BASE } from '@/config'

  const { createInventory, updateInventory, deleteInventory, fetchStock, createAdjustment } = useInventories()

  const {
    selected: selectedRows,
    deleting: bulkDeleting,
    error: selError,
    reset: resetSel,
    run: runBulk,
  } = useBulkDelete()

  function clearSelection () {
    resetSel()
    tableRef.value?.clearSelection?.()
  }

  async function bulkDeleteRows () {
    try {
      await runBulk('/inventories/bulk_delete', { ids: selectedRows.value })
      clearSelection()
      tableRef.value?.reload()
    } catch { /* error shown in the bar */ }
  }

  const apiURL = `${API_BASE}/inventories/`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Type', key: 'type', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Created At', key: 'created_at', align: 'end' },
    { title: '', key: 'actions', align: 'end', sortable: false, width: 56 },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const editingId = ref(null)

  const { handleSubmit, handleReset, setValues } = useForm({
    validationSchema: {
      name (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
      type (value) {
        if (editingId.value) return true
        return ['warehouse', 'store'].includes(value) || 'Select a type.'
      },
      code (value) {
        if (editingId.value) return true
        return (value?.length >= 1) || 'Code is required.'
      },
    },
  })

  const name = useField('name')
  const type = useField('type')
  const code = useField('code')
  const latitude = useField('latitude')
  const longitude = useField('longitude')

  function openCreate () {
    editingId.value = null
    handleReset()
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    setValues({ name: item.name, type: item.type, code: item.code })
    dialog.value = true
  }

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
    editingId.value = null
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      await (editingId.value
        ? updateInventory({ id: editingId.value, name: values.name })
        : createInventory({
          name: values.name,
          type: values.type,
          code: values.code,
          longitude: values.longitude || null,
          latitude: values.latitude || null,
        }))
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  const deleteDialog = ref(false)
  const deleteTarget = ref(null)
  const deleting = ref(false)
  const deleteError = ref('')

  function openDelete (item) {
    deleteTarget.value = item
    deleteError.value = ''
    deleteDialog.value = true
  }

  // Delete from inside the edit dialog: hand off to the confirm dialog.
  function deleteFromEdit () {
    if (!editingId.value) return
    const item = { id: editingId.value, name: name.value.value }
    dialog.value = false
    openDelete(item)
  }

  async function confirmDelete () {
    deleting.value = true
    deleteError.value = ''
    try {
      await deleteInventory(deleteTarget.value.id)
      deleteDialog.value = false
      tableRef.value?.reload()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }

  // --- Per-inventory stock view + quick "count" adjustment ---
  const stockDialog = ref(false)
  const stockTarget = ref(null)
  const stockRows = ref([])
  const stockError = ref('')
  const stockSearch = ref('')
  const adjustInputs = ref({})
  const adjustingId = ref(null)

  async function loadStock () {
    stockError.value = ''
    try {
      const { stock } = await fetchStock(stockTarget.value.id, { search: stockSearch.value || '' })
      stockRows.value = stock
    } catch (error) {
      stockError.value = error.message
    }
  }

  function openStock (item) {
    stockTarget.value = item
    stockRows.value = []
    adjustInputs.value = {}
    stockDialog.value = true
    loadStock()
  }

  let stockSearchTimer = null
  function onStockSearch () {
    clearTimeout(stockSearchTimer)
    stockSearchTimer = setTimeout(loadStock, 250)
  }

  async function applyCount (row) {
    const target = adjustInputs.value[row.variant_id]
    if (target === undefined || target === null || target === '') return
    adjustingId.value = row.variant_id
    stockError.value = ''
    try {
      await createAdjustment(stockTarget.value.id, {
        variantId: row.variant_id,
        mode: 'count',
        quantity: Number(target),
        note: 'manual count',
      })
      adjustInputs.value[row.variant_id] = null
      await loadStock()
    } catch (error) {
      stockError.value = error.message
    } finally {
      adjustingId.value = null
    }
  }
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
.inventory-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.inventory-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
:deep(.inventory-table tbody tr) {
  cursor: pointer;
}
</style>
