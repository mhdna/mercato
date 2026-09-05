<template>
  <v-container class="performance-page px-4 px-md-6 pb-4 pb-md-6 pt-1 pt-md-2" fluid>
    <h1 class="text-h4 font-weight-bold mb-5">Salesperson performance</h1>

    <v-alert class="formula-card mb-5" color="primary" icon="mdi-information" variant="tonal">
      <div class="font-weight-bold mb-1">Formula</div>
      <div class="text-body-2">
        Score = 50% × (sold amount ÷ highest amount) + 25% × (invoices ÷ highest invoices)
        + 25% × (goods sold ÷ highest goods)
      </div>
    </v-alert>

    <v-card class="mb-5" rounded="lg">
      <v-card-text>
        <div class="d-flex align-center ga-2 mb-4">
          <v-icon color="primary" icon="mdi-filter" />
          <span class="font-weight-bold">Filters</span>
          <v-spacer />
          <v-btn
            v-if="hasFilters"
            size="small"
            text="Clear all"
            variant="text"
            @click="clearFilters"
          />
        </div>
        <v-row density="compact">
          <v-col cols="12" lg="3" sm="6">
            <v-select
              v-model="branchId"
              clearable
              density="comfortable"
              hide-details
              item-title="name"
              item-value="id"
              :items="branches"
              label="Employee branch"
              prepend-inner-icon="mdi-store-outline"
              variant="outlined"
            />
          </v-col>
          <v-col cols="12" lg="3" sm="6">
            <v-autocomplete
              v-model="salespersonId"
              clearable
              density="comfortable"
              hide-details
              item-title="name"
              item-value="key"
              :items="availableSalespersons"
              label="Employee name"
              prepend-inner-icon="mdi-account-outline"
              variant="outlined"
            />
          </v-col>
          <v-col cols="12" lg="2" sm="6">
            <v-select
              v-model="period"
              density="comfortable"
              hide-details
              item-title="title"
              item-value="value"
              :items="periods"
              label="Period"
              prepend-inner-icon="mdi-calendar"
              variant="outlined"
            />
          </v-col>
          <v-col v-if="period === 'month'" cols="12" lg="2" sm="6">
            <v-menu v-model="monthMenu" :close-on-content-click="false">
              <template #activator="{ props }">
                <v-text-field
                  v-bind="props"
                  density="comfortable"
                  hide-details
                  label="Month"
                  :model-value="formattedMonth"
                  prepend-inner-icon="mdi-calendar"
                  readonly
                  variant="outlined"
                />
              </template>
              <v-card>
                <v-date-picker
                  v-model:month="monthPickerMonth"
                  v-model:year="monthPickerYear"
                  view-mode="months"
                  @update:month="selectMonth"
                />
              </v-card>
            </v-menu>
          </v-col>
          <v-col v-if="period === 'range'" cols="12" lg="4" sm="6">
            <v-menu v-model="rangeMenu" :close-on-content-click="false">
              <template #activator="{ props }">
                <v-text-field
                  v-bind="props"
                  density="comfortable"
                  hide-details
                  label="Date range"
                  :model-value="formattedDateRange"
                  prepend-inner-icon="mdi-calendar"
                  readonly
                  variant="outlined"
                />
              </template>
              <v-card>
                <v-date-picker v-model="dateRange" multiple="range" />
                <v-card-actions>
                  <v-spacer />
                  <v-btn text="Clear" variant="text" @click="dateRange = []" />
                  <v-btn color="primary" text="Done" variant="text" @click="rangeMenu = false" />
                </v-card-actions>
              </v-card>
            </v-menu>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-alert
      v-if="error"
      class="mb-5"
      closable
      type="error"
      variant="tonal"
      @click:close="error = ''"
    >
      {{ error }}
      <template #append><v-btn size="small" text="Retry" variant="text" @click="loadData" /></template>
    </v-alert>

    <v-row class="mb-3" density="compact">
      <v-col v-for="summary in summaries" :key="summary.label" cols="12" sm="4">
        <v-card class="h-100" rounded="lg">
          <v-card-text>
            <div class="text-caption text-medium-emphasis">{{ summary.label }}</div>
            <div class="text-h5 font-weight-bold mt-1">{{ summary.value }}</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-card class="mb-5" rounded="lg">
      <v-card-title class="d-flex align-center py-4">
        <span>Performance ranking</span>
        <v-spacer />
        <v-chip color="primary" size="small" variant="tonal">{{ rankedRows.length }} employees</v-chip>
      </v-card-title>
      <v-divider />
      <v-card-text v-if="loading" class="py-10 text-center"><v-progress-circular color="primary" indeterminate /></v-card-text>
      <v-card-text v-else-if="rankedRows.length === 0" class="py-10 text-center text-medium-emphasis">
        No salesperson activity matches these filters.
      </v-card-text>
      <v-card-text v-else class="pt-2">
        <div
          v-for="(employee, index) in rankedRows"
          :key="employee.key"
          class="performance-row py-4"
          role="button"
          tabindex="0"
          @click="openAdjustments(employee)"
          @keydown.enter="openAdjustments(employee)"
        >
          <div class="d-flex align-center ga-3 mb-3">
            <div class="rank text-caption font-weight-bold">{{ index + 1 }}</div>
            <v-avatar color="primary" size="38" variant="tonal">{{ initials(employee.name) }}</v-avatar>
            <div class="min-width-0">
              <div class="font-weight-bold text-truncate">{{ employee.name }}</div>
              <div class="text-caption text-medium-emphasis">{{ employee.branchName }}</div>
            </div>
            <v-spacer />
            <div class="text-end">
              <div class="text-h6 font-weight-bold">{{ employee.score.toFixed(1) }}</div>
              <div class="text-caption text-medium-emphasis">out of 100</div>
            </div>
          </div>
          <v-progress-linear :color="scoreColor(employee.score)" height="12" :model-value="employee.score" rounded />
          <div class="d-flex flex-wrap ga-2 mt-3">
            <v-chip size="small" variant="tonal">Amount {{ employee.amountScore.toFixed(0) }}%</v-chip>
            <v-chip size="small" variant="tonal">Invoices {{ employee.invoiceScore.toFixed(0) }}%</v-chip>
            <v-chip size="small" variant="tonal">Goods {{ employee.goodsScore.toFixed(0) }}%</v-chip>
            <v-chip color="warning" size="small" variant="tonal">
              Discounts {{ formatMoney(employee.totalDiscounts) }}
            </v-chip>
            <v-chip color="error" size="small" variant="tonal">
              Price overrides {{ formatMoney(employee.totalPriceOverrides) }}
            </v-chip>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <v-card rounded="lg">
      <v-card-title class="py-4">Performance details</v-card-title>
      <v-divider />
      <v-data-table
        class="performance-table"
        :headers="headers"
        :items="rankedRows"
        :loading="loading"
        no-data-text="No salesperson activity found"
        @click:row="openAdjustmentsFromTable"
      >
        <template #item.employee="{ item }">
          <div class="py-2">
            <div class="font-weight-medium">{{ item.name }}</div>
            <div class="text-caption text-medium-emphasis">{{ item.branchName }}</div>
          </div>
        </template>
        <template #item.score="{ item }">
          <v-chip :color="scoreColor(item.score)" size="small" variant="tonal">{{ item.score.toFixed(1) }}</v-chip>
        </template>
        <template #item.totalAmount="{ item }">{{ formatMoney(item.totalAmount) }}</template>
        <template #item.totalDiscounts="{ item }">{{ formatMoney(item.totalDiscounts) }}</template>
        <template #item.totalPriceOverrides="{ item }">{{ formatMoney(item.totalPriceOverrides) }}</template>
        <template #bottom />
      </v-data-table>
    </v-card>

    <v-dialog v-model="adjustmentsDialog" max-width="1100">
      <v-card class="adjustments-dialog-card" rounded="lg">
        <v-card-title class="d-flex align-center py-4">
          <div>
            <div>{{ selectedEmployee?.name }}</div>
            <div class="text-caption text-medium-emphasis">Manual discounts and price overrides</div>
          </div>
          <v-spacer />
          <v-btn icon="mdi-close" variant="text" @click="adjustmentsDialog = false" />
        </v-card-title>
        <v-divider />
        <v-card-text class="adjustments-dialog-content">
          <v-row class="mb-3" density="compact">
            <v-col cols="12" sm="6">
              <v-card color="warning" variant="tonal">
                <v-card-text>
                  <div class="text-caption">Total manual discounts</div>
                  <div class="text-h5 font-weight-bold">{{ formatMoney(selectedEmployee?.totalDiscounts) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="12" sm="6">
              <v-card color="error" variant="tonal">
                <v-card-text>
                  <div class="text-caption">Total price deductions from overrides</div>
                  <div class="text-h5 font-weight-bold">{{ formatMoney(selectedEmployee?.totalPriceOverrides) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <v-alert
            v-if="selectedAdjustments.length === 0"
            type="info"
            variant="tonal"
          >
            No manual discounts or price overrides were recorded for this salesperson in the selected period.
          </v-alert>
          <template v-else>
            <v-text-field
              v-model="adjustmentSearch"
              class="mb-3"
              clearable
              density="compact"
              hide-details
              label="Search adjustments"
              prepend-inner-icon="mdi-magnify"
              variant="outlined"
            />
            <v-data-table
              :headers="adjustmentHeaders"
              :items="selectedAdjustments"
              :items-per-page="14"
              :search="adjustmentSearch"
            >
              <template #item.type="{ item }">
                <v-chip :color="item.type === 'Discount' ? 'warning' : 'error'" size="small" variant="tonal">
                  {{ item.type }}
                </v-chip>
              </template>
              <template #item.date="{ item }">{{ formatDate(item.date) }}</template>
              <template #item.originalPrice="{ item }">{{ item.originalPrice == null ? '—' : formatMoney(item.originalPrice) }}</template>
              <template #item.finalPrice="{ item }">{{ formatMoney(item.finalPrice) }}</template>
              <template #item.amount="{ item }"><strong>{{ formatMoney(item.amount) }}</strong></template>
            </v-data-table>
          </template>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { authFetch } from '@/composables/useApi'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchInvoices } from '@/composables/useBranchInvoices'
  import { API_BASE } from '@/config'

  const { branches, fetchBranches } = useBranches()
  const { listBranchInvoiceItems } = useBranchInvoices()
  const invoices = ref([])
  const loading = ref(false)
  const error = ref('')
  const branchId = ref(null)
  const salespersonId = ref(null)
  const period = ref('month')
  const now = new Date()
  const month = ref(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)
  const monthMenu = ref(false)
  const monthPickerMonth = ref(now.getMonth())
  const monthPickerYear = ref(now.getFullYear())
  const dateRange = ref([])
  const rangeMenu = ref(false)
  const adjustmentsDialog = ref(false)
  const selectedEmployee = ref(null)
  const adjustmentSearch = ref('')

  const periods = [
    { title: 'This month', value: 'month' },
    { title: 'Custom date range', value: 'range' },
    { title: 'All time', value: 'all' },
  ]
  const headers = [
    { title: 'Employee', key: 'employee', align: 'start' },
    { title: 'Performance', key: 'score', align: 'center' },
    { title: 'Total Sold Amount', key: 'totalAmount', align: 'end' },
    { title: 'Total Invoices', key: 'totalInvoices', align: 'end' },
    { title: 'Total Number of goods', key: 'totalGoods', align: 'end' },
    { title: 'Total Discounts', key: 'totalDiscounts', align: 'end' },
    { title: 'Total Price Deductions', key: 'totalPriceOverrides', align: 'end' },
  ]
  const adjustmentHeaders = [
    { title: 'Type', key: 'type', align: 'start' },
    { title: 'Date', key: 'date', align: 'start' },
    { title: 'Invoice', key: 'invoiceCode', align: 'start' },
    { title: 'Product', key: 'product', align: 'start' },
    { title: 'Qty', key: 'quantity', align: 'end' },
    { title: 'Original Price', key: 'originalPrice', align: 'end' },
    { title: 'Final Price', key: 'finalPrice', align: 'end' },
    { title: 'Discount', key: 'percentage', align: 'end' },
    { title: 'Deducted Amount', key: 'amount', align: 'end' },
  ]

  function dateKey (value) {
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? '' : date.toISOString().slice(0, 10)
  }

  const formattedMonth = computed(() => new Date(`${month.value}-01T00:00:00`).toLocaleDateString(undefined, {
    month: 'long',
    year: 'numeric',
  }))
  const selectedRange = computed(() => dateRange.value.map(value => dateKey(value)).filter(Boolean).toSorted())
  const formattedDateRange = computed(() => {
    if (selectedRange.value.length === 0) return ''
    const format = value => new Date(`${value}T00:00:00`).toLocaleDateString()
    if (selectedRange.value.length === 1) return format(selectedRange.value[0])
    return `${format(selectedRange.value[0])} – ${format(selectedRange.value.at(-1))}`
  })

  function invoiceDate (invoice) {
    return new Date(invoice.occurred_at || invoice.created_at || invoice.received_at)
  }

  function inSelectedPeriod (invoice) {
    const date = invoiceDate(invoice)
    if (Number.isNaN(date.getTime())) return period.value === 'all'
    if (period.value === 'month') return date.toISOString().slice(0, 7) === month.value
    if (period.value === 'range') {
      const day = date.toISOString().slice(0, 10)
      const from = selectedRange.value[0]
      const to = selectedRange.value.at(-1)
      return (!from || day >= from) && (!to || day <= to)
    }
    return true
  }

  function personKey (invoice) {
    const id = invoice.salesperson_id ?? invoice.salesperson?.id
    return id == null ? `name:${invoice.salesperson_name || invoice.salesperson?.name || 'Unassigned'}` : String(id)
  }

  function goodsCount (invoice) {
    if (invoice.total_goods != null) return Number(invoice.total_goods) || 0
    if (invoice.items_count != null) return Number(invoice.items_count) || 0
    const items = invoice.items || invoice.invoice_items || []
    return items.reduce((total, item) => total + (Number(item.quantity) || 0), 0)
  }

  function manualDiscountAmount (item) {
    const source = item.discount_source || item.discount_type
    if (source && !['manual', 'salesperson', 'cashier'].includes(String(source).toLowerCase())) return 0
    if (!(Number(item.discount) > 0 || item.is_manual_discount || item.manual_discount)) return 0
    if (item.manual_discount_amount != null) return Math.max(0, Number(item.manual_discount_amount) || 0)
    const gross = (Number(item.unit_price) || 0) * Math.abs(Number(item.quantity) || 0)
    const net = Math.abs(Number(item.line_total) || 0)
    return Math.max(0, gross - net)
  }

  function originalUnitPrice (item) {
    const value = item.original_unit_price ?? item.list_price ?? item.regular_price ?? item.base_price
    return value == null ? null : Number(value)
  }

  function manualOverrideAmount (item) {
    if (item.price_override_amount != null) return Math.max(0, Number(item.price_override_amount) || 0)
    const source = item.price_override_source || item.override_source
    const isManual = item.is_manual_price_override || item.price_overridden || item.manual_price_override
      || ['manual', 'salesperson', 'cashier'].includes(String(source || '').toLowerCase())
    const original = originalUnitPrice(item)
    if (!isManual || original == null) return 0
    return Math.max(0, original - (Number(item.unit_price) || 0)) * Math.abs(Number(item.quantity) || 0)
  }

  function adjustmentRecords (invoice) {
    return (invoice.items || invoice.invoice_items || []).flatMap((item, index) => {
      const records = []
      const common = {
        id: `${invoice.id}-${item.id ?? index}`,
        date: invoice.occurred_at || invoice.created_at || invoice.received_at,
        invoiceCode: invoice.branch_invoice_code || invoice.invoice_code || `#${invoice.id}`,
        product: item.product_name || item.product_code || `Product #${item.branch_product_id ?? item.product_id ?? '—'}`,
        quantity: Math.abs(Number(item.quantity) || 0),
        finalPrice: Number(item.unit_price) || 0,
      }
      const discountAmount = manualDiscountAmount(item)
      if (discountAmount > 0) {
        records.push({
          ...common,
          id: `${common.id}-discount`,
          type: 'Discount',
          originalPrice: Number(item.unit_price) || 0,
          percentage: `${Number(item.discount) || 0}%`,
          amount: discountAmount,
        })
      }
      const overrideAmount = manualOverrideAmount(item)
      if (overrideAmount > 0) {
        records.push({
          ...common,
          id: `${common.id}-override`,
          type: 'Price override',
          originalPrice: originalUnitPrice(item),
          percentage: '—',
          amount: overrideAmount,
        })
      }
      return records
    })
  }

  const aggregatedRows = computed(() => {
    const result = new Map()
    for (const invoice of invoices.value) {
      if (invoice.kind && invoice.kind !== 'sale' && invoice.kind !== 'sales') continue
      if (branchId.value != null && String(invoice.branch_id) !== String(branchId.value)) continue
      if (!inSelectedPeriod(invoice)) continue
      const key = personKey(invoice)
      const row = result.get(key) || {
        key,
        name: invoice.salesperson_name || invoice.salesperson?.name || 'Unassigned',
        branchId: invoice.branch_id,
        totalAmount: 0,
        totalInvoices: 0,
        totalGoods: 0,
        totalDiscounts: 0,
        totalPriceOverrides: 0,
        adjustments: [],
      }
      row.totalAmount += Number(invoice.grand_total ?? invoice.total_amount ?? invoice.total ?? 0) || 0
      row.totalInvoices += 1
      row.totalGoods += goodsCount(invoice)
      const adjustments = adjustmentRecords(invoice)
      row.adjustments.push(...adjustments)
      row.totalDiscounts += adjustments.filter(item => item.type === 'Discount').reduce((sum, item) => sum + item.amount, 0)
      row.totalPriceOverrides += adjustments.filter(item => item.type === 'Price override').reduce((sum, item) => sum + item.amount, 0)
      result.set(key, row)
    }
    return [...result.values()].map(row => ({
      ...row,
      branchName: branches.value.find(branch => String(branch.id) === String(row.branchId))?.name || 'Main branch',
    }))
  })

  const availableSalespersons = computed(() => aggregatedRows.value.map(row => ({ key: row.key, name: row.name })))
  const rankedRows = computed(() => {
    const rows = salespersonId.value ? aggregatedRows.value.filter(row => row.key === salespersonId.value) : aggregatedRows.value
    const maxAmount = Math.max(...rows.map(row => row.totalAmount), 0)
    const maxInvoices = Math.max(...rows.map(row => row.totalInvoices), 0)
    const maxGoods = Math.max(...rows.map(row => row.totalGoods), 0)
    return rows.map(row => {
      const amountScore = maxAmount ? row.totalAmount / maxAmount * 100 : 0
      const invoiceScore = maxInvoices ? row.totalInvoices / maxInvoices * 100 : 0
      const goodsScore = maxGoods ? row.totalGoods / maxGoods * 100 : 0
      return { ...row, amountScore, invoiceScore, goodsScore, score: amountScore * 0.5 + invoiceScore * 0.25 + goodsScore * 0.25 }
    }).toSorted((a, b) => b.score - a.score)
  })

  const summaries = computed(() => [
    { label: 'Total sold amount', value: formatMoney(rankedRows.value.reduce((sum, row) => sum + row.totalAmount, 0)) },
    { label: 'Total invoices', value: rankedRows.value.reduce((sum, row) => sum + row.totalInvoices, 0).toLocaleString() },
    { label: 'Total goods sold', value: rankedRows.value.reduce((sum, row) => sum + row.totalGoods, 0).toLocaleString() },
  ])
  const selectedAdjustments = computed(() => selectedEmployee.value?.adjustments || [])
  const hasFilters = computed(() => branchId.value != null || salespersonId.value != null || period.value !== 'month' || month.value !== `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)

  function formatMoney (cents) {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format((Number(cents) || 0) / 100)
  }
  function formatDate (value) {
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? '—' : date.toLocaleString()
  }
  function initials (name) {
    return name.split(/\s+/).map(part => part[0]).slice(0, 2).join('').toUpperCase()
  }
  function scoreColor (score) {
    return score >= 75 ? 'success' : (score >= 45 ? 'warning' : 'error')
  }
  function selectMonth (monthIndex) {
    month.value = `${monthPickerYear.value}-${String(Number(monthIndex) + 1).padStart(2, '0')}`
    monthMenu.value = false
  }
  function openAdjustments (employee) {
    selectedEmployee.value = employee
    adjustmentSearch.value = ''
    adjustmentsDialog.value = true
  }
  function openAdjustmentsFromTable (_event, row) {
    openAdjustments(row.item)
  }
  function clearFilters () {
    branchId.value = null
    salespersonId.value = null
    period.value = 'month'
    month.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    dateRange.value = []
  }

  async function loadData () {
    loading.value = true
    error.value = ''
    try {
      const [, response] = await Promise.all([
        fetchBranches(),
        authFetch(`${API_BASE}/branch_invoices?page_size=100&page_id=0`),
      ])
      if (!response.ok) throw new Error(`Failed to load sales activity (${response.status})`)
      const data = await response.json()
      const loadedInvoices = Array.isArray(data) ? data : (data?.branch_invoices ?? [])
      const itemResults = await Promise.allSettled(loadedInvoices.map(invoice => {
        if (invoice.total_goods != null || invoice.items_count != null || invoice.items || invoice.invoice_items) {
          return Promise.resolve(invoice.items || invoice.invoice_items || [])
        }
        return listBranchInvoiceItems(invoice.id)
      }))
      invoices.value = loadedInvoices.map((invoice, index) => ({
        ...invoice,
        items: itemResults[index].status === 'fulfilled' ? itemResults[index].value : [],
      }))
    } catch (loadError) {
      error.value = loadError.message || 'Failed to load salesperson performance.'
      invoices.value = []
    } finally {
      loading.value = false
    }
  }

  watch(branchId, () => {
    salespersonId.value = null
  })
  watch(monthMenu, isOpen => {
    if (!isOpen) return
    const [year, selectedMonth] = month.value.split('-').map(Number)
    monthPickerYear.value = year
    monthPickerMonth.value = selectedMonth - 1
  })
  onMounted(loadData)
</script>

<style scoped>
  .performance-page { max-width: 1500px; }
  .formula-card { border-inline-start: 4px solid rgb(var(--v-theme-primary)); }
  .performance-row { cursor: pointer; border-radius: 8px; transition: background-color .15s ease; }
  .performance-row:hover, .performance-row:focus-visible { background: rgba(var(--v-theme-primary), .04); outline: none; }
  .performance-row + .performance-row { border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
  .performance-table :deep(tbody tr) { cursor: pointer; }
  .adjustments-dialog-card {
    display: flex;
    min-height: 660px;
    max-height: 660px !important;
    height: 660px;
    flex-direction: column;
  }
  .adjustments-dialog-content { flex: 1 1 auto; overflow-y: auto; }
  .rank { display: grid; place-items: center; width: 26px; height: 26px; border-radius: 50%; background: rgba(var(--v-theme-primary), .1); color: rgb(var(--v-theme-primary)); }
  .min-width-0 { min-width: 0; }
</style>
