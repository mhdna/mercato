<template>
  <v-dialog v-model="dialog" max-width="520">
    <v-card class="px-4">
      <v-card-title>Add a New Purchase</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-select
            v-model="supplierId.value.value"
            density="compact"
            :error-messages="supplierId.errorMessage.value"
            :item-title="s => s.name"
            :item-value="s => s.id"
            :items="suppliers"
            label="Supplier"
          />
          <v-select
            v-model="inventoryId.value.value"
            density="compact"
            :error-messages="inventoryId.errorMessage.value"
            :item-title="i => i.name"
            :item-value="i => i.id"
            :items="inventories"
            label="Destination Inventory"
          />
          <CurrencySelect
            v-model="currencyCode.value.value"
            :error-messages="currencyCode.errorMessage.value"
          />
          <v-text-field
            v-model="purchasedAt.value.value"
            density="compact"
            :error-messages="purchasedAt.errorMessage.value"
            label="Purchased At"
            type="date"
          />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" text="Create Purchase" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="itemsDialog" max-width="800">
    <v-card class="px-4">
      <v-card-title>
        Purchase #{{ itemsTarget?.id }}
        <v-chip class="ml-2" :color="statusColor(itemsTarget?.status)" size="small">{{ itemsTarget?.status }}</v-chip>
      </v-card-title>
      <v-card-text>
        <v-row v-if="itemsTarget?.status !== 'received'" align="center">
          <v-col cols="5">
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
          </v-col>
          <v-col cols="2">
            <v-text-field
              v-model.number="itemQuantity"
              density="compact"
              hide-details
              label="Quantity"
              type="number"
            />
          </v-col>
          <v-col cols="3">
            <v-text-field
              v-model.number="itemUnitPrice"
              density="compact"
              hide-details
              label="Unit Cost"
              type="number"
            />
          </v-col>
          <v-col cols="2">
            <v-btn :loading="itemSubmitting" text="Add" @click="addItem" />
          </v-col>
        </v-row>
        <v-alert v-if="itemError" class="mb-4 mt-2" type="error" variant="tonal">{{ itemError }}</v-alert>
        <v-table density="compact">
          <thead>
            <tr>
              <th>SKU</th>
              <th>Barcode</th>
              <th class="text-end">Quantity</th>
              <th class="text-end">Unit Cost</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ purchaseItemSku(item) }}</td>
              <td>{{ nullableString(item.variant_barcode) || '—' }}</td>
              <td class="text-end">{{ item.quantity }}</td>
              <td class="text-end">{{ item.unit_price }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions>
        <v-btn
          v-if="itemsTarget?.status === 'draft'"
          color="success"
          :loading="receiving"
          text="Receive"
          @click="receive"
        />
        <v-btn
          v-if="itemsTarget?.status === 'draft' || itemsTarget?.status === 'received'"
          color="error"
          :loading="cancelling"
          text="Cancel Purchase"
          variant="text"
          @click="cancel"
        />
        <v-spacer />
        <v-btn text="Close" @click="itemsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="purchase-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-invoice" />
        <span>Purchase Invoices</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="purchase-search"
          clearable
          density="compact"
          hide-details
          label="Search purchase invoices"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Purchase"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <ServerSideTable
        ref="tableRef"
        :api-u-r-l="apiURL"
        density="comfortable"
        :external-search="search"
        flush
        :headers="headers"
        hover
        root-key="purchases"
        :show-search-icon="false"
        @row-click="openItems"
      >
        <template #item.inventory_name="{ item }">
          {{ nullableString(item.inventory_name) || '—' }}
        </template>
        <template #item.status="{ item }">
          <v-chip :color="statusColor(item.status)" size="small">{{ item.status }}</v-chip>
        </template>
      </ServerSideTable>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { usePurchases } from '@/composables/usePurchases'
  import { useSuppliers } from '@/composables/useSuppliers'
  import { useVariants } from '@/composables/useVariants'
  import { API_BASE } from '@/config'

  const { createPurchase, fetchPurchase, addPurchaseItem, receivePurchase, cancelPurchase } = usePurchases()
  const { suppliers, fetchSuppliers } = useSuppliers()
  const { inventories, fetchInventories } = useInventories()
  const { variants: skuOptions, loading: skuLoading, searchVariants } = useVariants()
  fetchSuppliers()
  fetchInventories()
  searchVariants('')

  let skuSearchTimer = null
  function onSkuSearch (q) {
    clearTimeout(skuSearchTimer)
    skuSearchTimer = setTimeout(() => searchVariants(q || ''), 250)
  }

  const apiURL = `${API_BASE}/purchases`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Supplier', key: 'supplier_name', align: 'start' },
    { title: 'Inventory', key: 'inventory_name', align: 'start' },
    { title: 'Total', key: 'grand_total', align: 'end' },
    { title: 'Status', key: 'status', align: 'start' },
    { title: 'Purchased At', key: 'purchased_at', align: 'end' },
  ])

  function statusColor (s) {
    return { draft: 'grey', received: 'success', cancelled: 'error' }[s] || 'grey'
  }

  function nullableString (value) {
    if (typeof value === 'string') return value
    if (!value || typeof value !== 'object') return ''
    const valid = value.Valid ?? value.valid
    if (valid === false) return ''
    return value.String ?? value.string ?? ''
  }

  function purchaseItemSku (item) {
    const code = nullableString(item.product_code)
    const name = nullableString(item.product_name)
    return [code, name].filter(Boolean).join(' — ') || '—'
  }

  const tableRef = ref(null)
  const search = ref('')
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      supplierId (value) {
        return !!value || 'Select a supplier.'
      },
      inventoryId (value) {
        return !!value || 'Select a destination inventory.'
      },
      currencyCode (value) {
        return !!value || 'Select a currency.'
      },
      purchasedAt (value) {
        return !!value || 'Purchase date is required.'
      },
    },
  })

  const supplierId = useField('supplierId')
  const inventoryId = useField('inventoryId')
  const currencyCode = useField('currencyCode')
  const purchasedAt = useField('purchasedAt')

  function openCreate () {
    handleReset()
    dialog.value = true
  }

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const created = await createPurchase({
        supplierId: values.supplierId,
        inventoryId: values.inventoryId,
        currencyCode: values.currencyCode,
        purchasedAt: new Date(values.purchasedAt).toISOString(),
        items: [],
      })
      closeDialog()
      tableRef.value?.reload()
      openItems(created?.purchase ?? created)
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
  const itemUnitPrice = ref(null)
  const itemSubmitting = ref(false)
  const itemError = ref('')
  const receiving = ref(false)
  const cancelling = ref(false)

  async function openItems (item) {
    itemsTarget.value = item
    itemError.value = ''
    items.value = []
    itemsDialog.value = true
    await refreshItems()
  }

  async function refreshItems () {
    try {
      const data = await fetchPurchase(itemsTarget.value.id)
      itemsTarget.value = data?.purchase ?? itemsTarget.value
      items.value = data?.items ?? []
    } catch (error) {
      itemError.value = error.message
    }
  }

  async function addItem () {
    if (!itemVariantId.value || !itemQuantity.value) return
    itemSubmitting.value = true
    itemError.value = ''
    try {
      await addPurchaseItem({
        purchaseId: itemsTarget.value.id,
        variantId: itemVariantId.value,
        quantity: itemQuantity.value,
        unitPrice: itemUnitPrice.value || 0,
        currencyCode: itemsTarget.value.currency_code || 'USD',
      })
      itemVariantId.value = null
      itemQuantity.value = null
      itemUnitPrice.value = null
      await refreshItems()
    } catch (error) {
      itemError.value = error.message
    } finally {
      itemSubmitting.value = false
    }
  }

  async function receive () {
    receiving.value = true
    itemError.value = ''
    try {
      const res = await receivePurchase(itemsTarget.value.id)
      itemsTarget.value = res?.purchase ?? itemsTarget.value
      tableRef.value?.reload()
    } catch (error) {
      itemError.value = error.message
    } finally {
      receiving.value = false
    }
  }

  async function cancel () {
    cancelling.value = true
    itemError.value = ''
    try {
      const res = await cancelPurchase(itemsTarget.value.id)
      itemsTarget.value = res?.purchase ?? itemsTarget.value
      tableRef.value?.reload()
    } catch (error) {
      itemError.value = error.message
    } finally {
      cancelling.value = false
    }
  }
</script>

<style scoped>
.page-root {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.purchase-card {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.purchase-search {
  flex: 0 1 320px;
}
</style>
