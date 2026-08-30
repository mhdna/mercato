<template>
  <div class="d-flex ga-3 h-100" style="flex: 1 1 0; min-height: 0">
    <DashboardList
      ref="salesList"
      :api-url="salesUrl"
      icon="mdi-cart-check"
      :mock-items="salesMockItems"
      root-key="sales"
      style="flex: 1 1 0; min-width: 0"
      title="Sales"
    >
      <template #row="{ item }">
        <InitialsAvatar :name="item.client" :show-name="false" />
        <v-chip
          class="ms-2"
          :color="item.source === 'branch' ? 'primary' : 'secondary'"
          label
          size="x-small"
          variant="tonal"
        >
          {{ item.source }}
        </v-chip>
        <v-spacer />
        <span class="text-body-2 font-weight-medium me-2">{{ money(item.total) }}</span>
        <v-chip size="x-small" variant="tonal">{{ dateChip(item.occurred_at) }}</v-chip>
      </template>
      <template #details="{ item }">
        <DetailList
          :rows="[
            ['Source', item.source],
            ['Client', item.client],
            ['Amount', money(item.total)],
            ['Currency', item.currency],
            ['Date', dateFull(item.occurred_at)],
          ]"
        />
      </template>
    </DashboardList>

    <DashboardList
      ref="purchasesList"
      :api-url="purchasesUrl"
      icon="mdi-truck-delivery"
      :mock-items="purchasesMockItems"
      root-key="purchases"
      style="flex: 1 1 0; min-width: 0"
      title="Purchases"
    >
      <template #row="{ item }">
        <span class="text-body-2 text-truncate">{{ item.supplier }}</span>
        <v-spacer />
        <span class="text-body-2 font-weight-medium me-2">{{ money(item.total) }}</span>
        <v-chip size="x-small" variant="tonal">{{ dateChip(item.purchased_at) }}</v-chip>
      </template>
      <template #details="{ item }">
        <DetailList
          :rows="[
            ['Supplier', item.supplier],
            ['Items', item.items],
            ['Amount', money(item.total)],
            ['Currency', item.currency],
            ['Date', dateFull(item.purchased_at)],
          ]"
        />
      </template>
    </DashboardList>

    <DashboardList
      ref="expensesList"
      :api-url="expensesUrl"
      icon="mdi-receipt-text-check"
      :mock-items="expensesMockItems"
      root-key="expenses"
      style="flex: 1 1 0; min-width: 0"
      title="Recent Expenses"
    >
      <template #row="{ item }">
        <span class="text-body-2 text-truncate">{{ item.category || 'Uncategorised' }}</span>
        <v-spacer />
        <span class="text-body-2 font-weight-medium me-2">{{ money(item.amount) }}</span>
        <v-chip class="me-1" size="x-small" variant="tonal">{{ dateChip(item.occurred_at) }}</v-chip>
        <ExpenseReceiptThumb :image-id="item.image_id" />
      </template>
      <template #details="{ item }">
        <DetailList
          :rows="[
            ['Source', item.source],
            ['Description', item.description],
            ['Category', item.category],
            ['Amount', money(item.amount)],
            ['Currency', item.currency],
            ['Date', dateFull(item.occurred_at)],
          ]"
        />
      </template>
    </DashboardList>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { useDashboardLiveEvents } from '@/composables/useDashboardLiveEvents'
  import { API_BASE } from '@/config'
  import { useIncomeFilters } from './composables/useIncomeFilters'
  import DashboardList from './DashboardList.vue'
  import DetailList from './DetailList.vue'
  import ExpenseReceiptThumb from './ExpenseReceiptThumb.vue'
  import InitialsAvatar from './InitialsAvatar.vue'

  const { dashboardQuery } = useIncomeFilters()

  // Reactive: dashboardQuery() reads the shared filter refs, so each URL
  // recomputes on any filter change and DashboardList re-fetches from page 0.
  const salesUrl = computed(() => `${API_BASE}/dashboard/sales?${dashboardQuery()}`)
  const purchasesUrl = computed(() => `${API_BASE}/dashboard/purchases?${dashboardQuery()}`)
  const expensesUrl = computed(() => `${API_BASE}/dashboard/expenses?${dashboardQuery()}`)

  const salesMockItems = [
    { id: 1, client: 'Maya Haddad', source: 'branch', total: 18_450, currency: 'USD', occurred_at: '2026-08-28T09:20:00Z' },
    { id: 2, client: 'Omar Saad', source: 'central', total: 9275, currency: 'USD', occurred_at: '2026-08-27T15:45:00Z' },
    { id: 3, client: 'Lina Nassar', source: 'branch', total: 32_600, currency: 'USD', occurred_at: '2026-08-26T11:10:00Z' },
    { id: 4, client: 'Karim Tabet', source: 'branch', total: 7580, currency: 'USD', occurred_at: '2026-08-25T13:35:00Z' },
    { id: 5, client: 'Nour Khoury', source: 'central', total: 12_900, currency: 'USD', occurred_at: '2026-08-24T16:05:00Z' },
    { id: 6, client: 'Rami Daher', source: 'branch', total: 21_750, currency: 'USD', occurred_at: '2026-08-23T10:40:00Z' },
    { id: 7, client: 'Sara Mansour', source: 'branch', total: 6850, currency: 'USD', occurred_at: '2026-08-22T14:15:00Z' },
    { id: 8, client: 'Jad Karam', source: 'central', total: 15_300, currency: 'USD', occurred_at: '2026-08-21T12:50:00Z' },
    { id: 9, client: 'Dana Farah', source: 'branch', total: 28_425, currency: 'USD', occurred_at: '2026-08-20T09:30:00Z' },
    { id: 10, client: 'Tarek Saliba', source: 'central', total: 9950, currency: 'USD', occurred_at: '2026-08-19T17:20:00Z' },
    { id: 11, client: 'Rana Aoun', source: 'branch', total: 17_600, currency: 'USD', occurred_at: '2026-08-18T11:45:00Z' },
    { id: 12, client: 'Ziad Harb', source: 'branch', total: 24_100, currency: 'USD', occurred_at: '2026-08-17T15:10:00Z' },
  ]

  const purchasesMockItems = [
    { id: 1, supplier: 'Cedars Wholesale', items: 42, total: 28_600, currency: 'USD', purchased_at: '2026-08-28T08:05:00Z' },
    { id: 2, supplier: 'Beyrouth Trading', items: 18, total: 14_350, currency: 'USD', purchased_at: '2026-08-27T12:30:00Z' },
    { id: 3, supplier: 'North Star Supplies', items: 27, total: 19_475, currency: 'USD', purchased_at: '2026-08-25T10:15:00Z' },
    { id: 4, supplier: 'Green Valley Foods', items: 35, total: 22_100, currency: 'USD', purchased_at: '2026-08-23T16:40:00Z' },
    { id: 5, supplier: 'Levant Distributors', items: 24, total: 17_850, currency: 'USD', purchased_at: '2026-08-22T09:25:00Z' },
    { id: 6, supplier: 'Coastal Goods', items: 51, total: 34_200, currency: 'USD', purchased_at: '2026-08-21T13:50:00Z' },
    { id: 7, supplier: 'Bekaa Fresh Market', items: 31, total: 20_650, currency: 'USD', purchased_at: '2026-08-20T07:45:00Z' },
    { id: 8, supplier: 'Atlas Packaging', items: 16, total: 11_900, currency: 'USD', purchased_at: '2026-08-19T15:20:00Z' },
    { id: 9, supplier: 'Phoenicia Imports', items: 44, total: 29_750, currency: 'USD', purchased_at: '2026-08-18T11:35:00Z' },
    { id: 10, supplier: 'Metro Wholesale', items: 29, total: 18_300, currency: 'USD', purchased_at: '2026-08-17T14:05:00Z' },
    { id: 11, supplier: 'Summit Supplies', items: 38, total: 25_480, currency: 'USD', purchased_at: '2026-08-16T10:55:00Z' },
    { id: 12, supplier: 'Cedar Coast Foods', items: 22, total: 16_725, currency: 'USD', purchased_at: '2026-08-15T08:40:00Z' },
  ]

  const expensesMockItems = [
    { id: 1, source: 'branch', description: 'Store electricity bill', category: 'Utilities', amount: 8450, currency: 'USD', occurred_at: '2026-08-28T07:30:00Z', image_id: null },
    { id: 2, source: 'central', description: 'Delivery fuel', category: 'Transport', amount: 5200, currency: 'USD', occurred_at: '2026-08-27T14:20:00Z', image_id: null },
    { id: 3, source: 'branch', description: 'Packaging supplies', category: 'Supplies', amount: 6780, currency: 'USD', occurred_at: '2026-08-26T09:50:00Z', image_id: null },
    { id: 4, source: 'central', description: 'Equipment maintenance', category: 'Maintenance', amount: 11_200, currency: 'USD', occurred_at: '2026-08-24T11:05:00Z', image_id: null },
    { id: 5, source: 'branch', description: 'Internet subscription', category: 'Utilities', amount: 3200, currency: 'USD', occurred_at: '2026-08-23T08:35:00Z', image_id: null },
    { id: 6, source: 'central', description: 'Office cleaning service', category: 'Services', amount: 4750, currency: 'USD', occurred_at: '2026-08-22T16:15:00Z', image_id: null },
    { id: 7, source: 'branch', description: 'Printer paper and ink', category: 'Office', amount: 2650, currency: 'USD', occurred_at: '2026-08-21T10:20:00Z', image_id: null },
    { id: 8, source: 'central', description: 'Warehouse insurance', category: 'Insurance', amount: 14_500, currency: 'USD', occurred_at: '2026-08-20T12:40:00Z', image_id: null },
    { id: 9, source: 'branch', description: 'Refrigerator repair', category: 'Maintenance', amount: 8900, currency: 'USD', occurred_at: '2026-08-19T09:10:00Z', image_id: null },
    { id: 10, source: 'central', description: 'Social media campaign', category: 'Marketing', amount: 7200, currency: 'USD', occurred_at: '2026-08-18T13:25:00Z', image_id: null },
    { id: 11, source: 'branch', description: 'Staff transportation', category: 'Transport', amount: 4100, currency: 'USD', occurred_at: '2026-08-17T17:45:00Z', image_id: null },
    { id: 12, source: 'central', description: 'Security system service', category: 'Services', amount: 6350, currency: 'USD', occurred_at: '2026-08-16T11:30:00Z', image_id: null },
  ]

  const salesList = ref(null)
  const purchasesList = ref(null)
  const expensesList = ref(null)

  // Filter changes reset the lists via their reactive api-url; a branch
  // sync push has to nudge them.
  useDashboardLiveEvents(() => {
    salesList.value?.reset()
    purchasesList.value?.reset()
    expensesList.value?.reset()
  })

  function money (cents) {
    const n = Number(cents || 0) / 100
    return `${Number.isInteger(n) ? n : n.toFixed(2)}$`
  }
  function dateChip (value) {
    return value ? new Date(value).toLocaleDateString([], { day: '2-digit', month: 'short' }) : '—'
  }
  function dateFull (value) {
    return value ? new Date(value).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' }) : '—'
  }
</script>
