<template>
  <v-dialog v-model="itemsDialog" max-width="640">
    <v-card>
      <v-card-title>Invoice Items -- {{ itemsTarget?.branch_invoice_code }}</v-card-title>
      <v-card-text>
        <v-alert v-if="itemsError" class="mb-4" type="error" variant="tonal">{{ itemsError }}</v-alert>
        <v-table density="compact">
          <thead>
            <tr>
              <th>Branch Product ID</th>
              <th class="text-end">Qty</th>
              <th class="text-end">Unit Price</th>
              <th class="text-end">Discount</th>
              <th class="text-end">Line Total</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ item.branch_product_id }}</td>
              <td class="text-end">{{ item.quantity }}</td>
              <td class="text-end">{{ formatMoney(item.unit_price) }}</td>
              <td class="text-end">{{ item.discount }}%</td>
              <td class="text-end">{{ formatMoney(item.line_total) }}</td>
            </tr>
          </tbody>
        </v-table>
        <div v-if="itemsLoading" class="d-flex justify-center pa-4">
          <v-progress-circular color="primary" indeterminate />
        </div>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="itemsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Branch Invoices</h2>
    <v-select
      v-model="branchFilter"
      clearable
      density="compact"
      hide-details
      item-title="name"
      item-value="id"
      :items="branches"
      placeholder="All branches"
      style="max-width: 260px"
      variant="outlined"
      @update:model-value="onFilterChange"
    />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="buildApiUrl()"
    :headers="headers"
    root-key="branch_invoices"
  >
    <template #item.branch_id="{ item }">
      {{ branchName(item.branch_id) }}
    </template>
    <template #item.kind="{ item }">
      <v-chip :color="item.kind === 'return' ? 'warning' : 'primary'" size="small">{{ item.kind }}</v-chip>
    </template>
    <template #item.grand_total="{ item }">
      {{ formatMoney(item.grand_total) }}
    </template>
    <template #item.occurred_at="{ item }">
      {{ new Date(item.occurred_at).toLocaleString() }}
    </template>
    <template #item.received_at="{ item }">
      {{ new Date(item.received_at).toLocaleString() }}
    </template>
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-eye" size="small" variant="text" @click="openItems(item)" />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchInvoices } from '@/composables/useBranchInvoices'
  import { API_BASE } from '@/config'

  const { branches, fetchBranches } = useBranches()
  const { listBranchInvoiceItems } = useBranchInvoices()

  onMounted(() => fetchBranches())

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  function formatMoney (cents) {
    return (Number(cents) / 100).toFixed(2)
  }

  const headers = ref([
    { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
    { title: 'Kind', key: 'kind', align: 'start', sortable: false },
    { title: 'Code', key: 'branch_invoice_code', align: 'start', sortable: false },
    { title: 'Grand Total', key: 'grand_total', align: 'end', sortable: false },
    { title: 'Occurred At', key: 'occurred_at', align: 'start', sortable: false },
    { title: 'Received At', key: 'received_at', align: 'start', sortable: false },
    { title: '', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const branchFilter = ref(null)

  function buildApiUrl () {
    const url = new URL(`${API_BASE}/branch_invoices`, window.location.origin)
    if (branchFilter.value) url.searchParams.set('branch_id', branchFilter.value)
    return url.toString()
  }

  function onFilterChange () {
    tableRef.value?.reload()
  }

  const itemsDialog = ref(false)
  const itemsTarget = ref(null)
  const items = ref([])
  const itemsLoading = ref(false)
  const itemsError = ref('')

  async function openItems (invoice) {
    itemsTarget.value = invoice
    itemsDialog.value = true
    itemsLoading.value = true
    itemsError.value = ''
    items.value = []
    try {
      items.value = await listBranchInvoiceItems(invoice.id)
    } catch (error) {
      itemsError.value = error.message
    } finally {
      itemsLoading.value = false
    }
  }
</script>
