<template>
  <v-dialog v-model="dialog" max-width="480" :persistent="true">
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

  <v-dialog v-model="itemsDialog" max-width="720">
    <v-card class="px-4">
      <v-card-title>Items in Purchase #{{ itemsTarget?.id }}</v-card-title>
      <v-card-text>
        <v-row align="center">
          <v-col cols="4">
            <v-select
              v-model="itemProductId"
              density="compact"
              hide-details
              :item-title="p => `${p.code} - ${p.name}`"
              :item-value="p => p.id"
              :items="products"
              label="Product"
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
          <v-col cols="2">
            <v-text-field
              v-model.number="itemUnitPrice"
              density="compact"
              hide-details
              label="Unit Price"
              type="number"
            />
          </v-col>
          <v-col cols="3">
            <v-select
              v-model="itemCurrencyCode"
              density="compact"
              hide-details
              :item-title="c => c.code"
              :item-value="c => c.code"
              :items="currencies"
              label="Currency"
            />
          </v-col>
          <v-col cols="1">
            <v-btn :loading="itemSubmitting" text="Add" @click="addItem" />
          </v-col>
        </v-row>
        <v-alert v-if="itemError" class="mb-4 mt-2" type="error" variant="tonal">{{ itemError }}</v-alert>
        <v-table density="compact">
          <thead>
            <tr>
              <th>Product ID</th>
              <th>Quantity</th>
              <th>Unit Price</th>
              <th>Currency</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ item.product_id }}</td>
              <td>{{ item.quantity }}</td>
              <td>{{ item.unit_price }}</td>
              <td>{{ item.currency_code }}</td>
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
    <h2 class="text-h6">Purchase Invoices</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Purchase" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="purchases"
  >
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-format-list-bulleted" size="small" variant="text" @click="openItems(item)" />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useCurrencies } from '@/composables/useCurrencies'
  import { useProducts } from '@/composables/useProducts'
  import { usePurchases } from '@/composables/usePurchases'
  import { useSuppliers } from '@/composables/useSuppliers'
  import { API_BASE } from '@/config'

  const { createPurchase, addPurchaseItem } = usePurchases()
  const { suppliers, fetchSuppliers } = useSuppliers()
  const { products, fetchProducts } = useProducts()
  const { currencies, fetchCurrencies } = useCurrencies()
  fetchSuppliers()
  fetchProducts()
  fetchCurrencies()

  const apiURL = `${API_BASE}/purchases`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Supplier ID', key: 'supplier_id', align: 'start' },
    { title: 'Purchased At', key: 'purchased_at', align: 'end' },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      supplierId (value) {
        return !!value || 'Select a supplier.'
      },
      purchasedAt (value) {
        return !!value || 'Purchase date is required.'
      },
    },
  })

  const supplierId = useField('supplierId')
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
        supplier_id: values.supplierId,
        purchased_at: new Date(values.purchasedAt).toISOString(),
      })
      closeDialog()
      tableRef.value?.reload()
      openItems(created)
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
  const itemUnitPrice = ref(null)
  const itemCurrencyCode = ref(null)
  const itemSubmitting = ref(false)
  const itemError = ref('')

  function openItems (item) {
    itemsTarget.value = item
    itemError.value = ''
    items.value = []
    itemsDialog.value = true
  }

  async function addItem () {
    if (!itemProductId.value || !itemQuantity.value || !itemUnitPrice.value || !itemCurrencyCode.value) return
    itemSubmitting.value = true
    itemError.value = ''
    try {
      const item = await addPurchaseItem({
        purchaseId: itemsTarget.value.id,
        productId: itemProductId.value,
        quantity: itemQuantity.value,
        unitPrice: itemUnitPrice.value,
        currencyCode: itemCurrencyCode.value,
      })
      items.value.push(item)
      itemProductId.value = null
      itemQuantity.value = null
      itemUnitPrice.value = null
    } catch (error) {
      itemError.value = error.message
    } finally {
      itemSubmitting.value = false
    }
  }
</script>
