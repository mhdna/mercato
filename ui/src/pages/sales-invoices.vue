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
                @update:model-value="onCashboxChange"
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
                v-model="salespersonId.value.value"
                clearable
                density="compact"
                item-title="name"
                item-value="id"
                :items="salespersons"
                label="Salesperson (optional)"
                :loading="salespersonsLoading"
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
                  <v-autocomplete
                    v-model="line.variantId"
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

  <v-dialog v-model="detailsDialog" max-width="720">
    <v-card>
      <v-card-title>Invoice Details -- {{ details?.invoice?.invoice_code }}</v-card-title>
      <v-card-text>
        <v-alert v-if="detailsError" class="mb-4" type="error" variant="tonal">{{ detailsError }}</v-alert>
        <div v-if="detailsLoading" class="d-flex justify-center pa-4">
          <v-progress-circular color="primary" indeterminate />
        </div>
        <template v-else-if="details">
          <v-row density="compact">
            <v-col cols="6" md="4"><span class="text-caption">Kind</span><div class="text-capitalize">{{ details.kind }}</div></v-col>
            <v-col cols="6" md="4"><span class="text-caption">Client</span><div>{{ details.client?.name || '--' }}</div></v-col>
            <v-col cols="6" md="4"><span class="text-caption">Salesperson</span><div>{{ details.salesperson?.name || '--' }}</div></v-col>
            <v-col cols="6" md="4"><span class="text-caption">Loyalty Points</span><div>{{ details.invoice?.loyalty_points_delta }}</div></v-col>
            <v-col cols="6" md="4"><span class="text-caption">Grand Total</span><div>{{ formatMoney(details.invoice?.grand_total) }}</div></v-col>
            <v-col v-if="details.related_sales_invoice_id" cols="6" md="4">
              <span class="text-caption">Original Sale</span><div>#{{ details.related_sales_invoice_id }}</div>
            </v-col>
          </v-row>

          <v-divider class="my-3" />
          <div class="text-subtitle-2 mb-1">Items</div>
          <v-table density="compact">
            <thead>
              <tr>
                <th>Product</th>
                <th class="text-end">Qty</th>
                <th class="text-end">Unit Price</th>
                <th class="text-end">Discount</th>
                <th class="text-end">Line Total</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in details.items" :key="item.product_id">
                <td>{{ item.product_code }} -- {{ item.product_name }}</td>
                <td class="text-end">{{ item.quantity }}</td>
                <td class="text-end">{{ formatMoney(item.unit_price) }}</td>
                <td class="text-end">{{ item.discount }}%</td>
                <td class="text-end">{{ formatMoney(item.line_total) }}</td>
              </tr>
            </tbody>
          </v-table>

          <div class="text-subtitle-2 mt-3 mb-1">Settlement</div>
          <v-table density="compact">
            <thead>
              <tr>
                <th>Account</th>
                <th class="text-end">Amount</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="payment in details.payments" :key="payment.id">
                <td>{{ payment.account_name }}</td>
                <td class="text-end">{{ formatMoney(payment.amount) }}</td>
              </tr>
              <tr v-if="!details.payments?.length">
                <td class="text-medium-emphasis" colspan="2">No settlement recorded.</td>
              </tr>
            </tbody>
          </v-table>
        </template>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="detailsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <InvoiceTypesDialog v-model="invoiceTypesDialog" />

  <div class="page-root">
    <v-card class="invoices-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-invoice" />
        <span>Invoices</span>
        <v-spacer />
        <template v-if="invoiceTab === 'retail'">
          <v-text-field
            v-model="retailSearch"
            class="invoice-search"
            clearable
            density="compact"
            hide-details
            label="Search retail invoices"
            prepend-inner-icon="mdi-magnify"
            variant="outlined"
          />
          <v-select
            v-model="retailBranchFilter"
            clearable
            density="compact"
            hide-details
            item-title="name"
            item-value="id"
            :items="branches"
            placeholder="All branches"
            style="max-width: 260px"
            variant="outlined"
          />
        </template>
        <template v-else>
          <v-text-field
            v-model="wholesaleSearch"
            class="invoice-search"
            clearable
            density="compact"
            hide-details
            label="Search wholesale invoices"
            prepend-inner-icon="mdi-magnify"
            variant="outlined"
          />
          <v-btn
            prepend-icon="mdi-tag-multiple"
            text="Invoice Types"
            variant="tonal"
            @click="invoiceTypesDialog = true"
          />
          <v-btn
            color="primary"
            prepend-icon="mdi-plus"
            text="Add Invoice"
            variant="flat"
            @click="openCreate"
          />
        </template>
      </v-card-title>
      <v-divider />

      <div class="invoices-layout">
        <v-tabs
          v-model="invoiceTab"
          class="invoice-tabs"
          color="primary"
          direction="vertical"
        >
          <v-tab prepend-icon="mdi-store-outline" text="Retail" value="retail" />
          <v-tab prepend-icon="mdi-cart" text="Wholesale" value="wholesale" />
        </v-tabs>

        <v-divider vertical />

        <div class="invoices-content">
          <div v-show="invoiceTab === 'retail'">
            <BranchInvoices
              :branch-filter="retailBranchFilter"
              :branches="branches"
              :search="retailSearch"
              :search-debounce="100"
            />
          </div>

          <div v-show="invoiceTab === 'wholesale'" class="pa-4">
            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="apiURL"
              density="comfortable"
              :external-search="wholesaleSearch"
              flush
              :headers="headers"
              root-key="sales_invoices"
              :search-debounce="100"
              :show-search-icon="false"
            >
              <template #item.actions="{ item }">
                <v-icon-btn icon="mdi-eye" size="small" variant="text" @click="openDetails(item)" />
              </template>
            </ServerSideTable>
          </div>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { computed, ref } from 'vue'
  import InvoiceTypesDialog from '@/components/InvoiceTypesDialog.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBranches } from '@/composables/useBranches'
  import { useCashboxAccounts } from '@/composables/useCashboxAccounts'
  import { useCashboxes } from '@/composables/useCashboxes'
  import { useClients } from '@/composables/useClients'
  import { useInventories } from '@/composables/useInventories'
  import { useInvoiceDetails } from '@/composables/useInvoiceDetails'
  import { usePosSettings } from '@/composables/usePosSettings'
  import { usePriceLists } from '@/composables/usePriceLists'
  import { useSalesInvoices } from '@/composables/useSalesInvoices'
  import { useShifts } from '@/composables/useShifts'
  import { useVariants } from '@/composables/useVariants'
  import { API_BASE } from '@/config'
  import BranchInvoices from '@/pages/branch-invoices.vue'

  const { branches, fetchBranches } = useBranches()
  const { createSalesInvoice } = useSalesInvoices()
  const { cashboxes, fetchCashboxes } = useCashboxes()
  const { cashboxAccounts, fetchCashboxAccounts } = useCashboxAccounts()
  const { shifts, fetchShifts } = useShifts()
  const { inventories, fetchInventories } = useInventories()
  const { clients, fetchClients } = useClients()
  const { priceLists, fetchPriceLists } = usePriceLists()
  const { variants: skuOptions, loading: skuLoading, searchVariants } = useVariants()

  let skuSearchTimer = null
  function onSkuSearch (q) {
    clearTimeout(skuSearchTimer)
    skuSearchTimer = setTimeout(() => searchVariants(q || ''), 250)
  }
  const { getPosSettings } = usePosSettings()
  const { getInvoiceDetails } = useInvoiceDetails()

  fetchCashboxes()
  fetchBranches()
  fetchCashboxAccounts()
  fetchShifts()
  fetchInventories()
  fetchClients()
  fetchPriceLists()
  searchVariants('')

  const openShifts = computed(() => shifts.value.filter(s => !s.is_closed))

  const apiURL = `${API_BASE}/sales_invoices`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Invoice Code', key: 'invoice_code', align: 'start' },
    { title: 'Client ID', key: 'client_id', align: 'start' },
    { title: 'Subtotal', key: 'subtotal', align: 'end' },
    { title: 'Grand Total', key: 'grand_total', align: 'end' },
    { title: 'Created At', key: 'created_at', align: 'end' },
    { title: '', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const invoiceTypesDialog = ref(false)
  const invoiceTab = ref('retail')
  const retailSearch = ref('')
  const retailBranchFilter = ref(null)
  const wholesaleSearch = ref('')
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
  const salespersonId = useField('salespersonId')
  const discount = useField('discount')

  // Salespersons are scoped to a cashbox (see salespersons.cashbox_id) --
  // reuses the same /pos_settings reference-data endpoint kashi-pos itself
  // uses for this dropdown, rather than a new one.
  const salespersons = ref([])
  const salespersonsLoading = ref(false)

  async function onCashboxChange (cashboxIdValue) {
    salespersonId.value.value = null
    salespersons.value = []
    if (!cashboxIdValue) return
    salespersonsLoading.value = true
    try {
      const data = await getPosSettings(cashboxIdValue)
      salespersons.value = data?.salespersons ?? []
    } catch {
      salespersons.value = []
    } finally {
      salespersonsLoading.value = false
    }
  }

  let lineKey = 0
  const lines = ref([])

  function addLine () {
    lines.value.push({ key: lineKey++, variantId: null, quantity: 1, unitPrice: 0, discount: 0 })
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

  function formatMoney (cents) {
    return (Number(cents ?? 0) / 100).toFixed(2)
  }

  function openCreate () {
    handleReset()
    discount.value.value = 0
    lines.value = []
    salespersons.value = []
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
        variant_id: l.variantId,
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
        salesperson_id: values.salespersonId || null,
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

  const detailsDialog = ref(false)
  const details = ref(null)
  const detailsLoading = ref(false)
  const detailsError = ref('')

  async function openDetails (invoice) {
    detailsDialog.value = true
    detailsLoading.value = true
    detailsError.value = ''
    details.value = null
    try {
      details.value = await getInvoiceDetails(invoice.id)
    } catch (error) {
      detailsError.value = error.message
    } finally {
      detailsLoading.value = false
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
.invoices-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.invoices-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.invoice-tabs {
  flex: 0 0 160px;
}
.invoices-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.invoice-search {
  flex: 0 1 320px;
}
</style>
