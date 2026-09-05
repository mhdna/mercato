<template>
  <v-dialog v-model="dialog" max-width="560">
    <v-card class="px-4">
      <v-card-title>{{ editingId ? 'Edit Transfer' : 'Add a New Transfer' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-row>
            <v-col cols="6">
              <v-select
                v-model="fromInventoryId.value.value"
                density="compact"
                :error-messages="fromInventoryId.errorMessage.value"
                :item-title="i => i.name"
                :item-value="i => i.id"
                :items="inventories"
                label="From Inventory"
              />
            </v-col>
            <v-col cols="6">
              <v-select
                v-model="toInventoryId.value.value"
                density="compact"
                :error-messages="toInventoryId.errorMessage.value"
                :item-title="i => i.name"
                :item-value="i => i.id"
                :items="inventories"
                label="To Inventory"
              />
            </v-col>
          </v-row>
          <v-text-field v-model="note.value.value" density="compact" label="Note (optional)" />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingId ? 'Save' : 'Create Transfer'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="itemsDialog" max-width="720">
    <v-card class="px-4">
      <v-card-title>
        Transfer #{{ itemsTarget?.id }}
        <v-chip class="ml-2" :color="statusColor(itemsTarget?.status)" size="small">{{ itemsTarget?.status }}</v-chip>
      </v-card-title>
      <v-card-text>
        <div v-if="itemsTarget?.status === 'draft'" class="d-flex align-center mb-4" style="gap: 12px;">
          <v-autocomplete
            v-model="itemVariantId"
            density="compact"
            hide-details
            hide-no-data
            :item-title="v => v.label"
            :item-value="v => v.id"
            :items="skuOptions"
            :loading="skuLoading"
            no-filter
            placeholder="Search SKU / barcode"
            @update:search="onSkuSearch"
          />
          <v-text-field
            v-model.number="itemQuantity"
            density="compact"
            hide-details
            label="Quantity"
            style="max-width: 140px;"
            type="number"
          />
          <v-btn :loading="itemSubmitting" text="Add" @click="addItem" />
        </div>
        <v-alert v-if="itemError" class="mb-4" type="error" variant="tonal">{{ itemError }}</v-alert>
        <v-table density="compact">
          <thead>
            <tr>
              <th>SKU</th>
              <th>Barcode</th>
              <th class="text-end">Quantity</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ item.product_code }} — {{ item.product_name }}</td>
              <td>{{ item.variant_barcode }}</td>
              <td class="text-end">{{ item.quantity }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions>
        <v-btn
          v-if="itemsTarget?.status === 'draft'"
          color="primary"
          :loading="staging"
          text="Dispatch"
          @click="stage('dispatch')"
        />
        <v-btn
          v-if="itemsTarget?.status === 'dispatched'"
          color="success"
          :loading="staging"
          text="Receive"
          @click="stage('receive')"
        />
        <v-spacer />
        <v-btn text="Close" @click="itemsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="transfers-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-transfer" />
        <span>Transfers</span>
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Transfer"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <div class="transfers-content">
        <div class="pa-4">
          <ServerSideTable
            ref="tableRef"
            :api-u-r-l="apiURL"
            density="comfortable"
            flush
            :headers="headers"
            :max-page-size="10"
            root-key="transfers"
            :show-search-icon="false"
          >
            <template #item.status="{ item }">
              <v-chip :color="statusColor(item.status)" size="small">{{ item.status }}</v-chip>
            </template>
            <template #item.actions="{ item }">
              <v-icon-btn icon="mdi-format-list-bulleted" size="small" variant="text" @click="openItems(item)" />
              <v-icon-btn
                v-if="item.status === 'draft'"
                icon="mdi-pencil"
                size="small"
                variant="text"
                @click="openEdit(item)"
              />
            </template>
          </ServerSideTable>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { useTransfers } from '@/composables/useTransfers'
  import { useVariants } from '@/composables/useVariants'
  import { API_BASE } from '@/config'

  const {
    createTransfer, updateTransfer, fetchTransferItems, createTransferItem,
    dispatchTransfer, receiveTransfer,
  } = useTransfers()
  const { inventories, fetchInventories } = useInventories()
  const { variants: skuOptions, loading: skuLoading, searchVariants } = useVariants()
  fetchInventories()
  searchVariants('')

  let skuSearchTimer = null
  function onSkuSearch (q) {
    clearTimeout(skuSearchTimer)
    skuSearchTimer = setTimeout(() => searchVariants(q || ''), 250)
  }

  const apiURL = `${API_BASE}/transfers`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'From', key: 'from_inventory_name', align: 'start' },
    { title: 'To', key: 'to_inventory_name', align: 'start' },
    { title: 'Items', key: 'item_count', align: 'end' },
    { title: 'Status', key: 'status', align: 'start' },
    { title: 'Created At', key: 'created_at', align: 'end' },
    { title: '', key: 'actions', align: 'end', sortable: false, width: 96 },
  ])

  function statusColor (s) {
    return { draft: 'grey', dispatched: 'warning', received: 'success', cancelled: 'error' }[s] || 'grey'
  }

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const editingId = ref(null)

  const { handleSubmit, handleReset, setValues } = useForm({
    validationSchema: {
      fromInventoryId (value) {
        return !!value || 'Select the source inventory.'
      },
      toInventoryId (value, { form }) {
        if (!value) return 'Select the destination inventory.'
        if (value === form.fromInventoryId) return 'From and To inventories must be different.'
        return true
      },
    },
  })

  const fromInventoryId = useField('fromInventoryId')
  const toInventoryId = useField('toInventoryId')
  const note = useField('note')

  function openCreate () {
    editingId.value = null
    handleReset()
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    setValues({ fromInventoryId: item.from_inventory_id, toInventoryId: item.to_inventory_id, note: item.note })
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
        ? updateTransfer({
          id: editingId.value,
          from_inventory_id: values.fromInventoryId,
          to_inventory_id: values.toInventoryId,
          type: 'products',
          note: values.note || '',
        })
        : createTransfer({
          fromInventoryId: values.fromInventoryId,
          toInventoryId: values.toInventoryId,
          note: values.note || '',
          items: [],
        }))
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  const itemsDialog = ref(false)
  const itemsTarget = ref(null)
  const items = ref([])
  const itemVariantId = ref(null)
  const itemQuantity = ref(null)
  const itemSubmitting = ref(false)
  const itemError = ref('')
  const staging = ref(false)

  async function openItems (item) {
    itemsTarget.value = item
    itemError.value = ''
    itemsDialog.value = true
    try {
      items.value = await fetchTransferItems(item.id)
    } catch (error) {
      itemError.value = error.message
    }
  }

  async function addItem () {
    if (!itemVariantId.value || !itemQuantity.value) return
    itemSubmitting.value = true
    itemError.value = ''
    try {
      await createTransferItem({
        transferId: itemsTarget.value.id,
        variantId: itemVariantId.value,
        quantity: itemQuantity.value,
      })
      items.value = await fetchTransferItems(itemsTarget.value.id)
      itemVariantId.value = null
      itemQuantity.value = null
    } catch (error) {
      itemError.value = error.message
    } finally {
      itemSubmitting.value = false
    }
  }

  async function stage (action) {
    staging.value = true
    itemError.value = ''
    try {
      const fn = action === 'dispatch' ? dispatchTransfer : receiveTransfer
      const res = await fn(itemsTarget.value.id)
      itemsTarget.value = res?.transfer ?? itemsTarget.value
      tableRef.value?.reload()
    } catch (error) {
      itemError.value = error.message
    } finally {
      staging.value = false
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
.transfers-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.transfers-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
</style>
