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

  <v-dialog v-model="itemsDialog" max-width="640">
    <v-card class="px-4">
      <v-card-title>Items in Transfer #{{ itemsTarget?.id }}</v-card-title>
      <v-card-text>
        <div class="d-flex align-center mb-4" style="gap: 12px;">
          <v-select
            v-model="itemProductId"
            density="compact"
            hide-details
            :item-title="p => `${p.code} - ${p.name}`"
            :item-value="p => p.id"
            :items="products"
            label="Product"
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
              <th>Product ID</th>
              <th>Quantity</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ item.product_id }}</td>
              <td>{{ item.quantity }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="itemsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Transfers</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Transfer" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="transfers"
  >
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-format-list-bulleted" size="small" variant="text" @click="openItems(item)" />
      <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEdit(item)" />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { useProducts } from '@/composables/useProducts'
  import { useTransfers } from '@/composables/useTransfers'
  import { API_BASE } from '@/config'

  const { createTransfer, updateTransfer, fetchTransferItems, createTransferItem } = useTransfers()
  const { inventories, fetchInventories } = useInventories()
  const { products, fetchProducts } = useProducts()
  fetchInventories()
  fetchProducts()

  const apiURL = `${API_BASE}/transfers`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'From Inventory', key: 'from_inventory_id', align: 'start' },
    { title: 'To Inventory', key: 'to_inventory_id', align: 'start' },
    { title: 'Type', key: 'type', align: 'start' },
    { title: 'Created At', key: 'created_at', align: 'end' },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

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

  function openCreate () {
    editingId.value = null
    handleReset()
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    setValues({ fromInventoryId: item.from_inventory_id, toInventoryId: item.to_inventory_id })
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
        })
        : createTransfer({
          from_inventory_id: values.fromInventoryId,
          to_inventory_id: values.toInventoryId,
          type: 'products',
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
  const itemProductId = ref(null)
  const itemQuantity = ref(null)
  const itemSubmitting = ref(false)
  const itemError = ref('')

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
    if (!itemProductId.value || !itemQuantity.value) return
    itemSubmitting.value = true
    itemError.value = ''
    try {
      await createTransferItem({
        transferId: itemsTarget.value.id,
        productId: itemProductId.value,
        quantity: itemQuantity.value,
      })
      items.value = await fetchTransferItems(itemsTarget.value.id)
      itemProductId.value = null
      itemQuantity.value = null
    } catch (error) {
      itemError.value = error.message
    } finally {
      itemSubmitting.value = false
    }
  }
</script>
