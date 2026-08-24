<template>
  <v-dialog v-model="dialog" max-width="960" scrollable>
    <v-card class="px-4">
      <v-card-title>New Sales Invoice</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-row>
            <v-col cols="6" md="3">
              <v-select
                v-model="cashboxId.value.value"
                density="compact"
                :error-messages="cashboxId.errorMessage.value"
                :item-title="c => c.name"
                :item-value="c => c.id"
                :items="cashboxes"
                label="Cashbox"
              />
            </v-col>
            <v-col cols="6" md="3">
              <v-select
                v-model="cashboxAccountId.value.value"
                density="compact"
                :error-messages="cashboxAccountId.errorMessage.value"
                :item-title="a => a.name"
                :item-value="a => a.id"
                :items="cashboxAccounts"
                label="Cashbox Account"
              />
            </v-col>
            <v-col cols="6" md="3">
              <v-select
                v-model="shiftId.value.value"
                density="compact"
                :error-messages="shiftId.errorMessage.value"
                :item-title="s => `#${s.id} (cashbox ${s.cashbox_id})`"
                :item-value="s => s.id"
                :items="openShifts"
                label="Shift"
              />
            </v-col>
            <v-col cols="6" md="3">
              <v-select
                v-model="inventoryId.value.value"
                density="compact"
                :error-messages="inventoryId.errorMessage.value"
                :item-title="i => i.name"
                :item-value="i => i.id"
                :items="inventories"
                label="Inventory"
              />
            </v-col>
            <v-col cols="6" md="3">
              <v-select
                v-model="clientId.value.value"
                density="compact"
                :error-messages="clientId.errorMessage.value"
                :item-title="c => c.name"
                :item-value="c => c.id"
                :items="clients"
                label="Client"
              />
            </v-col>
            <v-col cols="6" md="3">
              <v-select
                v-model="priceListId.value.value"
                clearable
                density="compact"
                :item-title="p => p.name"
                :item-value="p => p.id"
                :items="priceLists"
                label="Price List (optional)"
              />
            </v-col>
            <v-col cols="6" md="3">
              <v-text-field
                v-model.number="discount.value.value"
                density="compact"
                :error-messages="discount.errorMessage.value"
                label="Invoice Discount %"
                type="number"
              />
            </v-col>
          </v-row>

          <v-divider class="my-2" />
          <div class="d-flex justify-space-between align-center mb-2">
            <span class="text-subtitle-2">Line Items</span>
            <v-btn
              prepend-icon="mdi-plus"
              size="small"
              text="Add Line"
              variant="tonal"
              @click="addLine"
            />
          </div>

          <v-table density="compact">
            <thead>
              <tr>
                <th>Product</th>
                <th style="width: 110px;">Quantity</th>
                <th style="width: 130px;">Unit Price</th>
                <th style="width: 110px;">Discount %</th>
                <th style="width: 110px;">Line Total</th>
                <th style="width: 40px;" />
              </tr>
            </thead>
            <tbody>
              <tr v-for="(line, idx) in lines" :key="line.key">
                <td>
                  <v-select
                    v-model="line.productId"
                    density="compact"
                    hide-details
                    :item-title="p => `${p.code} - ${p.name}`"
                    :item-value="p => p.id"
                    :items="products"
                  />
                </td>
                <td><v-text-field v-model.number="line.quantity" density="compact" hide-details type="number" /></td>
                <td><v-text-field v-model.number="line.unitPrice" density="compact" hide-details type="number" /></td>
                <td><v-text-field v-model.number="line.discount" density="compact" hide-details type="number" /></td>
                <td class="text-end">{{ lineTotal(line) }}</td>
                <td>
                  <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="lines.splice(idx, 1)" />
                </td>
              </tr>
            </tbody>
          </v-table>

          <v-row class="mt-2">
            <v-col class="text-end" cols="4">Items Total: <strong>{{ itemsTotal }}</strong></v-col>
            <v-col class="text-end" cols="4">Sub Total: <strong>{{ subTotal }}</strong></v-col>
            <v-col class="text-end" cols="4">Grand Total: <strong>{{ grandTotal }}</strong></v-col>
          </v-row>

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn
              color="primary"
              :disabled="lines.length === 0"
              :loading="submitting"
              text="Create Invoice"
              type="submit"
            />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Sales Invoices</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Invoice" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="sales_invoices"
  />
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { computed, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useCashboxAccounts } from '@/composables/useCashboxAccounts'
  import { useCashboxes } from '@/composables/useCashboxes'
  import { useClients } from '@/composables/useClients'
  import { useInventories } from '@/composables/useInventories'
  import { usePriceLists } from '@/composables/usePriceLists'
  import { useProducts } from '@/composables/useProducts'
  import { useSalesInvoices } from '@/composables/useSalesInvoices'
  import { useShifts } from '@/composables/useShifts'
  import { API_BASE } from '@/config'

  const { createSalesInvoice } = useSalesInvoices()
  const { cashboxes, fetchCashboxes } = useCashboxes()
  const { cashboxAccounts, fetchCashboxAccounts } = useCashboxAccounts()
  const { shifts, fetchShifts } = useShifts()
  const { inventories, fetchInventories } = useInventories()
  const { clients, fetchClients } = useClients()
  const { priceLists, fetchPriceLists } = usePriceLists()
  const { products, fetchProducts } = useProducts()

  fetchCashboxes()
  fetchCashboxAccounts()
  fetchShifts()
  fetchInventories()
  fetchClients()
  fetchPriceLists()
  fetchProducts()

  const openShifts = computed(() => shifts.value.filter(s => !s.is_closed))

  const apiURL = `${API_BASE}/sales_invoices`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Invoice Code', key: 'invoice_code', align: 'start' },
    { title: 'Client ID', key: 'client_id', align: 'start' },
    { title: 'Subtotal', key: 'subtotal', align: 'end' },
    { title: 'Grand Total', key: 'grand_total', align: 'end' },
    { title: 'Created At', key: 'created_at', align: 'end' },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      cashboxId (value) {
        return !!value || 'Select a cashbox.'
      },
      cashboxAccountId (value) {
        return !!value || 'Select a cashbox account.'
      },
      shiftId (value) {
        return !!value || 'Select an open shift.'
      },
      inventoryId (value) {
        return !!value || 'Select an inventory.'
      },
      clientId (value) {
        return !!value || 'Select a client.'
      },
      discount () {
        return true
      },
    },
  })

  const cashboxId = useField('cashboxId')
  const cashboxAccountId = useField('cashboxAccountId')
  const shiftId = useField('shiftId')
  const inventoryId = useField('inventoryId')
  const clientId = useField('clientId')
  const priceListId = useField('priceListId')
  const discount = useField('discount')

  let lineKey = 0
  const lines = ref([])

  function addLine () {
    lines.value.push({ key: lineKey++, productId: null, quantity: 1, unitPrice: 0, discount: 0 })
  }

  function itemPrice (line) {
    return (Number(line.unitPrice) || 0) * (Number(line.quantity) || 0)
  }

  function lineTotal (line) {
    const price = itemPrice(line)
    const d = Number(line.discount) || 0
    return d > 0 ? price - Math.floor(price * d / 100) : price
  }

  const itemsTotal = computed(() => lines.value.reduce((sum, l) => sum + itemPrice(l), 0))
  const subTotal = computed(() => {
    const invoiceDiscount = Number(discount.value.value) || 0
    return Math.floor(itemsTotal.value * (100 - invoiceDiscount) / 100)
  })
  const grandTotal = computed(() => lines.value.reduce((sum, l) => sum + lineTotal(l), 0))

  function openCreate () {
    handleReset()
    discount.value.value = 0
    lines.value = []
    addLine()
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
      const items = lines.value.map(l => ({
        product_id: l.productId,
        unit_price: Number(l.unitPrice) || 0,
        line_total: lineTotal(l),
        discount: Number(l.discount) || 0,
        quantity: Number(l.quantity) || 0,
      }))

      await createSalesInvoice({
        cashbox_id: values.cashboxId,
        cashbox_account_id: values.cashboxAccountId,
        shift_id: values.shiftId,
        inventory_id: values.inventoryId,
        year: new Date().getFullYear(),
        client_id: values.clientId,
        discount: Number(values.discount) || 0,
        sub_total: subTotal.value,
        discounted_total: itemsTotal.value,
        grand_total: grandTotal.value,
        price_list_id: values.priceListId || null,
        items,
      })
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })
</script>
