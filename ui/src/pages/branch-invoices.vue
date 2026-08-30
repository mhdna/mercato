<template>
  <v-dialog v-model="itemsDialog" max-width="760">
    <v-card class="invoice-details-card">
      <v-card-title class="d-flex align-center ga-3 px-6 py-4">
        <v-avatar color="primary" size="38" variant="tonal">
          <v-icon icon="mdi-invoice" size="22" />
        </v-avatar>
        <div>
          <div class="text-subtitle-1 font-weight-bold">Invoice details</div>
          <div class="text-caption text-medium-emphasis">{{ itemsTarget?.branch_invoice_code }}</div>
        </div>
      </v-card-title>
      <v-divider />
      <v-card-text class="invoice-details-content pa-6">
        <v-alert v-if="itemsError" class="mb-4" type="error" variant="tonal">{{ itemsError }}</v-alert>
        <div v-if="itemsLoading" class="d-flex justify-center pa-4">
          <v-progress-circular color="primary" indeterminate />
        </div>
        <template v-else-if="itemsDetails">
          <v-sheet class="pa-0" rounded="lg">
            <v-row dense>
              <v-col cols="6" md="4">
                <span class="text-caption">Branch</span>
                <div>{{ branchName(itemsDetails.invoice?.branch_id) }}</div>
              </v-col>
              <v-col cols="6" md="4">
                <span class="text-caption">Kind</span>
                <div class="text-capitalize">{{ itemsDetails.invoice?.kind }}</div>
              </v-col>
              <v-col cols="6" md="4">
                <span class="text-caption">Salesperson</span>
                <div>{{ itemsDetails.invoice?.salesperson_name || '--' }}</div>
              </v-col>
              <v-col cols="6" md="4">
                <span class="text-caption">Loyalty Points</span>
                <div>{{ itemsDetails.invoice?.loyalty_points_delta }}</div>
              </v-col>
              <v-col cols="6" md="4">
                <span class="text-caption">Grand Total</span>
                <div class="font-weight-bold">{{ formatMoney(itemsDetails.invoice?.grand_total) }}</div>
              </v-col>
              <v-col v-if="nullableString(itemsDetails.invoice?.related_client_ref)" cols="6" md="4">
                <span class="text-caption">Original Sale Ref</span>
                <div class="text-truncate">{{ nullableString(itemsDetails.invoice.related_client_ref) }}</div>
              </v-col>
            </v-row>
          </v-sheet>

          <div class="text-subtitle-2 mt-6 mb-2">Items</div>
          <v-table class="details-table" density="comfortable">
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
              <tr v-for="item in itemsDetails.items" :key="item.id">
                <td>{{ item.branch_product_id }}</td>
                <td class="text-end">{{ item.quantity }}</td>
                <td class="text-end">{{ formatMoney(item.unit_price) }}</td>
                <td class="text-end">{{ item.discount }}%</td>
                <td class="text-end">{{ formatMoney(item.line_total) }}</td>
              </tr>
            </tbody>
          </v-table>

          <div class="text-subtitle-2 mt-6 mb-2">Settlement</div>
          <v-table class="details-table" density="comfortable">
            <thead>
              <tr>
                <th>Account</th>
                <th class="text-end">Amount</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="payment in itemsDetails.payments" :key="payment.id">
                <td>{{ payment.account_name }}</td>
                <td class="text-end">{{ formatMoney(payment.amount) }}</td>
              </tr>
              <tr v-if="!itemsDetails.payments?.length">
                <td class="text-medium-emphasis" colspan="2">No settlement reported (older kashi-pos build).</td>
              </tr>
            </tbody>
          </v-table>
        </template>
      </v-card-text>
      <v-divider />
      <v-card-actions class="px-6 py-4">
        <v-spacer />
        <v-btn
          v-if="itemsTarget?.kind === 'sales'"
          color="primary"
          text="Return"
          variant="tonal"
          @click="returnViewedInvoice"
        />
        <v-btn text="Close" @click="itemsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    density="comfortable"
    :external-search="search"
    flush
    :headers="headers"
    hover
    :query-params="queryParams"
    root-key="branch_invoices"
    :search-debounce="searchDebounce"
    :show-search-icon="false"
    @row-click="openItems"
  >
    <template #item.branch_id="{ item }">
      {{ branchName(item.branch_id) }}
    </template>
    <template #item.kind="{ item }">
      <v-chip :color="kindColor(item.kind)" size="small">{{ item.kind }}</v-chip>
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
  </ServerSideTable>

  <v-dialog v-model="remoteReturnDialog" max-width="560">
    <v-card>
      <v-card-title>Return remotely -- {{ remoteReturnTarget?.branch_invoice_code }}</v-card-title>
      <v-card-text>
        <template v-if="remoteReturnPhase === 'items'">
          <v-alert class="mb-4" type="info" variant="tonal">
            This queues a command for {{ branchName(remoteReturnTarget?.branch_id) }} to process
            the return itself. It isn't executed instantly -- the branch picks it up next time
            it's online.
          </v-alert>
          <v-alert v-if="remoteReturnError" class="mb-4" type="error" variant="tonal">{{ remoteReturnError }}</v-alert>

          <div v-if="remoteReturnItemsLoading" class="d-flex justify-center pa-4">
            <v-progress-circular color="primary" indeterminate />
          </div>
          <template v-else>
            <div class="text-body-2 mb-2">Items to return</div>
            <v-table density="compact">
              <thead>
                <tr>
                  <th />
                  <th>Branch Product ID</th>
                  <th class="text-end">Qty</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in remoteReturnRows" :key="row.branch_product_id">
                  <td>
                    <v-checkbox-btn v-model="row.selected" density="compact" />
                  </td>
                  <td>{{ row.branch_product_id }}</td>
                  <td class="text-end">
                    <v-text-field
                      v-model.number="row.quantity"
                      density="compact"
                      :disabled="!row.selected"
                      hide-details
                      :max="row.maxQuantity"
                      min="1"
                      style="max-width: 90px"
                      type="number"
                    />
                  </td>
                </tr>
              </tbody>
            </v-table>
          </template>

          <v-textarea v-model="remoteReturnReason" class="mt-4" label="Reason (optional)" rows="2" />
        </template>

        <div v-else-if="remoteReturnPhase === 'polling'" class="d-flex flex-column align-center pa-6">
          <v-progress-circular class="mb-4" color="primary" indeterminate />
          <div class="text-body-2 text-center">
            Waiting for {{ branchName(remoteReturnTarget?.branch_id) }} to process the return...
          </div>
        </div>

        <template v-else-if="remoteReturnPhase === 'result'">
          <v-alert
            :text="remoteReturnResult.status === 'success'
              ? 'Return processed successfully.'
              : remoteReturnResult.status === 'failed'
                ? remoteReturnResult.error || 'The branch reported a failure with no error message.'
                : 'Still pending -- the branch may be offline right now. It will process the next time it\'s online, and this list will update live once it does.'"
            :type="remoteReturnResult.status === 'success' ? 'success' : remoteReturnResult.status === 'failed' ? 'error' : 'info'"
            variant="tonal"
          />
        </template>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          v-if="remoteReturnPhase === 'items'"
          text="Cancel"
          @click="closeRemoteReturn"
        />
        <v-btn
          v-if="remoteReturnPhase === 'items'"
          color="primary"
          :disabled="!hasSelectedReturnItems"
          :loading="remoteReturnLoading"
          text="Queue return"
          @click="submitRemoteReturn"
        />
        <v-btn v-else text="Close" @click="closeRemoteReturn" />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { useBranchInvoices } from '@/composables/useBranchInvoices'
  import { useBranchSettings } from '@/composables/useBranchSettings'
  import { useInvoiceDetails } from '@/composables/useInvoiceDetails'
  import { API_BASE } from '@/config'

  const props = defineProps({
    branches: { type: Array, default: () => [] },
    branchFilter: { type: [Number, String], default: null },
    search: { type: String, default: '' },
    searchDebounce: { type: Number, default: 100 },
  })

  const { listBranchInvoiceItems, requestRemoteReturn } = useBranchInvoices()
  const { listBranchCommands } = useBranchSettings()
  const { getBranchInvoiceDetails } = useInvoiceDetails()

  // Live refresh -- api/branch_sync.go broadcasts "branch_invoice_created"
  // for every kind (sales/return/exchange), so a remotely-triggered return
  // (see requestRemoteReturn) shows up here as soon as the branch actually
  // processes and syncs it back, same as any branch-initiated invoice.
  const { ensureConnected, onMessage } = useAdminSocket()
  let unsubscribe = null
  onMounted(() => {
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (message.type !== 'branch_invoice_created') return
      if (props.branchFilter && message.branch_id !== props.branchFilter) return
      tableRef.value?.reload()
    })
  })
  onUnmounted(() => unsubscribe?.())

  function branchName (id) {
    return props.branches.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  function kindColor (kind) {
    if (kind === 'return') return 'warning'
    if (kind === 'exchange') return 'info'
    return 'primary'
  }

  function formatMoney (cents) {
    return (Number(cents) / 100).toFixed(2)
  }

  function nullableString (value) {
    if (typeof value === 'string') return value
    if (!value || typeof value !== 'object') return ''
    const valid = value.Valid ?? value.valid
    if (valid === false) return ''
    return value.String ?? value.string ?? ''
  }

  const headers = ref([
    { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
    { title: 'Kind', key: 'kind', align: 'start', sortable: false },
    { title: 'Code', key: 'branch_invoice_code', align: 'start', sortable: false },
    { title: 'Grand Total', key: 'grand_total', align: 'end', sortable: false },
    { title: 'Occurred At', key: 'occurred_at', align: 'start', sortable: false },
    { title: 'Received At', key: 'received_at', align: 'start', sortable: false },
  ])

  const apiURL = `${API_BASE}/branch_invoices`
  const tableRef = ref(null)
  const queryParams = computed(() => ({
    branch_id: props.branchFilter || null,
  }))

  const itemsDialog = ref(false)
  const itemsTarget = ref(null)
  const itemsDetails = ref(null)
  const itemsLoading = ref(false)
  const itemsError = ref('')

  async function openItems (invoice) {
    itemsTarget.value = invoice
    itemsDialog.value = true
    itemsLoading.value = true
    itemsError.value = ''
    itemsDetails.value = null
    try {
      itemsDetails.value = await getBranchInvoiceDetails(invoice.id)
    } catch (error) {
      itemsError.value = error.message
    } finally {
      itemsLoading.value = false
    }
  }

  function returnViewedInvoice () {
    const invoice = itemsTarget.value
    itemsDialog.value = false
    openRemoteReturn(invoice)
  }

  const remoteReturnDialog = ref(false)
  const remoteReturnTarget = ref(null)
  const remoteReturnReason = ref('')
  const remoteReturnLoading = ref(false)
  const remoteReturnError = ref('')
  const remoteReturnRows = ref([])
  const remoteReturnItemsLoading = ref(false)
  // 'items' (picking what to return) -> 'polling' (command queued, waiting
  // on the branch) -> 'result' (final success/failed/still-pending state).
  const remoteReturnPhase = ref('items')
  const remoteReturnResult = ref(null)
  let remoteReturnPollToken = 0

  const hasSelectedReturnItems = computed(() => remoteReturnRows.value.some(row => row.selected))

  async function openRemoteReturn (invoice) {
    remoteReturnTarget.value = invoice
    remoteReturnReason.value = ''
    remoteReturnError.value = ''
    remoteReturnRows.value = []
    remoteReturnPhase.value = 'items'
    remoteReturnDialog.value = true
    remoteReturnItemsLoading.value = true
    try {
      const invoiceItems = await listBranchInvoiceItems(invoice.id)
      // Every line starts selected at its full quantity, so leaving
      // everything checked reproduces today's whole-invoice return.
      remoteReturnRows.value = invoiceItems.map(item => ({
        branch_product_id: item.branch_product_id,
        maxQuantity: item.quantity,
        quantity: item.quantity,
        selected: true,
      }))
    } catch (error) {
      remoteReturnError.value = error.message
    } finally {
      remoteReturnItemsLoading.value = false
    }
  }

  async function submitRemoteReturn () {
    remoteReturnLoading.value = true
    remoteReturnError.value = ''
    try {
      const items = remoteReturnRows.value
        .filter(row => row.selected)
        .map(row => ({
          branch_product_id: row.branch_product_id,
          quantity: Math.min(Math.max(1, Number(row.quantity) || 1), row.maxQuantity),
        }))
      const command = await requestRemoteReturn(remoteReturnTarget.value.branch_id, remoteReturnTarget.value, remoteReturnReason.value, items)
      remoteReturnPhase.value = 'polling'
      pollRemoteReturnCommand(remoteReturnTarget.value.branch_id, command?.id)
    } catch (error) {
      remoteReturnError.value = error.message
    } finally {
      remoteReturnLoading.value = false
    }
  }

  // The branch usually executes within a second or two if it's online (see
  // requestRemoteReturn), so a handful of quick polls covers the common
  // case; if it's still pending after that, it just hasn't come back
  // online yet -- the branch-invoices table's own WS listener (above) will
  // reflect it live whenever that happens, no need to keep polling forever.
  async function pollRemoteReturnCommand (branchId, commandId, attempt = 0) {
    const token = ++remoteReturnPollToken
    if (!commandId) {
      remoteReturnResult.value = { status: 'pending' }
      remoteReturnPhase.value = 'result'
      return
    }
    let commands
    try {
      commands = await listBranchCommands(branchId)
    } catch {
      commands = []
    }
    if (token !== remoteReturnPollToken) return // dialog closed/reopened meanwhile
    const command = commands.find(cmd => cmd.id === commandId)
    const status = command?.status ?? 'pending'
    if (status !== 'pending' || attempt >= 9) {
      remoteReturnResult.value = { status, error: command?.error?.Valid ? command.error.String : '' }
      remoteReturnPhase.value = 'result'
      return
    }
    setTimeout(() => pollRemoteReturnCommand(branchId, commandId, attempt + 1), 1500)
  }

  function closeRemoteReturn () {
    remoteReturnPollToken++ // abort any poll chain still in flight
    remoteReturnDialog.value = false
  }
</script>

<style scoped>
.invoice-search {
  flex: 0 1 320px;
}

.invoice-details-card {
  display: flex;
  height: 720px !important;
  max-height: calc(100vh - 48px);
  flex-direction: column;
}

.invoice-details-content {
  flex: 1 1 auto;
  overflow-y: auto;
}

.details-table {
  border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  border-radius: 8px;
}
</style>
