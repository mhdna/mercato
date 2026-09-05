<template>
  <div class="page-root">
    <v-alert
      v-if="loadError"
      class="mb-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ loadError }}
    </v-alert>

    <v-card class="loans-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-cash-multiple" />
        <span>Loans</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search loans"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-select
          v-model="statusFilter"
          clearable
          density="compact"
          hide-details
          :items="STATUS_OPTIONS"
          placeholder="Any status"
          style="max-width: 160px"
          variant="outlined"
        />
        <v-btn-toggle
          v-model="tab"
          color="primary"
          density="comfortable"
          mandatory
          variant="outlined"
        >
          <v-btn value="own">Central</v-btn>
          <v-btn value="branch">Branch</v-btn>
        </v-btn-toggle>
        <v-btn
          v-if="tab === 'own'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Loan"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <div class="loans-layout">
        <CategorySidebar
          v-model="selectedCategory"
          :categories="loanCategories"
          @add="openCreateCategory"
          @delete="removeCategory"
          @edit="openEditCategory"
        />

        <v-divider vertical />

        <div class="loans-content">
          <div v-if="tab === 'branch'" class="px-4 pt-3">
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
            />
          </div>

          <div class="pa-4">
            <BulkDeleteBar
              :count="selectedLoans.length"
              :error="loanSelError"
              :loading="loanDeleting"
              @clear="resetLoanSel"
              @confirm="bulkDeleteLoans"
            />
            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="`${API_BASE}/loans`"
              class="loans-table"
              density="comfortable"
              expandable
              :expanded="expanded"
              :external-search="search ?? ''"
              flush
              :headers="headers"
              :query-params="queryParams"
              root-key="loans"
              selectable
              :show-search-icon="false"
              @loaded="onRowsLoaded"
              @update:expanded="onExpandedChange"
              @update:selected="v => (selectedLoans = v)"
            >
              <template #item.data-table-select="{ item, internalItem, isSelected, toggleSelect }">
                <v-checkbox-btn
                  :disabled="item.origin !== 'central_loan'"
                  :model-value="isSelected(internalItem)"
                  @click.stop
                  @update:model-value="toggleSelect(internalItem)"
                />
              </template>
              <template #item.branch_id="{ item }">{{ branchName(item.branch_id) }}</template>
              <template #item.category_id="{ item }">
                <CategoryChip :category="categoryFor(item.category_id)" />
              </template>
              <template #item.amount="{ item }">
                {{ formatMoney(item.amount) }} {{ item.currency_code }}
              </template>
              <template #item.paid_amount="{ item }">
                {{ formatMoney(item.paid_amount) }}
                <span v-if="item.status !== 'paid'" class="text-medium-emphasis text-caption">
                  ({{ formatMoney(remaining(item)) }} left)
                </span>
              </template>
              <template #item.status="{ item }">
                <v-chip :color="STATUS_META[item.status]?.color" size="small" variant="tonal">
                  {{ STATUS_META[item.status]?.label ?? item.status }}
                </v-chip>
              </template>
              <template #item.occurred_at="{ item }">{{ formatDate(item.occurred_at) }}</template>
              <template #item.actions="{ item }">
                <v-icon-btn
                  v-if="item.origin === 'central_loan'"
                  icon="mdi-pencil"
                  size="small"
                  variant="text"
                  @click.stop="openEdit(item)"
                />
              </template>

              <template #expanded-row="{ columns, item }">
                <tr class="loans-expanded">
                  <td :colspan="columns.length">
                    <div class="py-3 px-2">
                      <div class="payments-header mb-3">
                        <div class="d-flex align-center flex-wrap ga-3">
                          <span class="text-h6 font-weight-bold">Payments</span>
                          <v-chip
                            :color="STATUS_META[item.status]?.color"
                            size="large"
                            variant="tonal"
                          >
                            <span class="font-weight-bold">{{ formatMoney(item.paid_amount) }}</span>
                            <span class="mx-1 text-medium-emphasis">/</span>
                            <span>{{ formatMoney(item.amount) }}</span>
                            <span class="ms-1 text-medium-emphasis">{{ item.currency_code }}</span>
                          </v-chip>
                        </div>
                        <v-progress-linear
                          class="mt-2"
                          :color="STATUS_META[item.status]?.color"
                          height="6"
                          :model-value="paidPercent(item)"
                          rounded
                        />
                      </div>

                      <v-alert
                        v-if="paymentsFor(item.id).error"
                        class="mb-2"
                        density="compact"
                        type="error"
                        variant="tonal"
                      >
                        {{ paymentsFor(item.id).error }}
                      </v-alert>

                      <v-table v-if="paymentsFor(item.id).list.length > 0" class="mb-3" density="compact">
                        <thead>
                          <tr>
                            <th class="text-start">Amount</th>
                            <th class="text-start">Note</th>
                            <th class="text-start">Paid At</th>
                            <th />
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="p in paymentsFor(item.id).list" :key="p.id">
                            <td>{{ formatMoney(p.amount) }} {{ p.currency_code }}</td>
                            <td>{{ p.note || '—' }}</td>
                            <td>{{ formatDate(p.paid_at) }}</td>
                            <td class="text-end">
                              <v-icon-btn
                                icon="mdi-delete"
                                size="x-small"
                                variant="text"
                                @click="removePayment(item, p)"
                              />
                            </td>
                          </tr>
                        </tbody>
                      </v-table>
                      <div
                        v-else-if="!paymentsFor(item.id).loading"
                        class="text-medium-emphasis text-caption mb-3"
                      >
                        No payments recorded yet.
                      </div>

                      <form
                        v-if="item.status !== 'paid'"
                        class="d-flex flex-wrap align-start ga-2"
                        @submit.prevent="addPayment(item)"
                      >
                        <v-text-field
                          v-model.number="payFormFor(item.id).amount"
                          density="compact"
                          :error-messages="payAmountError(item)"
                          :hint="`Remaining: ${formatMoney(remaining(item))} ${item.currency_code}`"
                          label="Amount"
                          persistent-hint
                          style="max-width: 180px"
                          type="number"
                          variant="outlined"
                        />
                        <CurrencySelect
                          v-model="payFormFor(item.id).currencyCode"
                          style="max-width: 130px"
                        />
                        <v-text-field
                          v-model="payFormFor(item.id).note"
                          density="compact"
                          hide-details
                          label="Note"
                          style="max-width: 220px"
                          variant="outlined"
                        />
                        <v-btn
                          color="primary"
                          :disabled="!!payAmountError(item)"
                          :loading="payFormFor(item.id).submitting"
                          text="Add Payment"
                          type="submit"
                          variant="flat"
                        />
                      </form>
                    </div>
                  </td>
                </tr>
              </template>
            </ServerSideTable>
          </div>
        </div>
      </div>
    </v-card>

    <!-- add / edit loan -->
    <v-dialog v-model="dialog" max-width="480">
      <v-card class="px-4">
        <v-card-title>{{ editingId ? 'Edit Loan' : 'Add Loan' }}</v-card-title>
        <v-card-text>
          <form @submit.prevent="submit">
            <v-text-field
              v-model="form.description"
              density="compact"
              :error-messages="formErrors.description"
              label="Description"
            />
            <v-select
              v-model="form.categoryId"
              density="compact"
              :error-messages="formErrors.categoryId"
              item-title="name"
              item-value="id"
              :items="activeCategories"
              label="Category"
            />
            <v-text-field
              v-model.number="form.amount"
              density="compact"
              :error-messages="formErrors.amount"
              label="Amount"
              type="number"
            />
            <CurrencySelect
              v-model="form.currencyCode"
              :error-messages="formErrors.currencyCode"
            />

            <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
              {{ submitError }}
            </v-alert>

            <v-card-actions class="px-0">
              <v-btn
                v-if="editingId"
                color="error"
                prepend-icon="mdi-delete"
                text="Delete"
                variant="text"
                @click="removeLoan"
              />
              <v-spacer />
              <v-btn text="Cancel" @click="closeDialog" />
              <v-btn
                color="primary"
                :loading="submitting"
                :text="editingId ? 'Save' : 'Add Loan'"
                type="submit"
              />
            </v-card-actions>
          </form>
        </v-card-text>
      </v-card>
    </v-dialog>

    <CategoryDialog
      v-model="categoryDialog"
      :category="editingCategory"
      name-hint="Lender source, e.g. Owner, External"
      :on-submit="submitCategory"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import CategoryChip from '@/components/CategoryChip.vue'
  import CategorySidebar from '@/components/CategorySidebar.vue'
  import CategoryDialog from '@/components/Forms/CategoryDialog.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { authFetch } from '@/composables/useApi'
  import { useBranches } from '@/composables/useBranches'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useLoanCategories } from '@/composables/useLoanCategories'
  import { API_BASE } from '@/config'

  const STATUS_META = {
    paid: { color: 'success', label: 'Paid' },
    partial: { color: 'warning', label: 'Partial' },
    unpaid: { color: 'grey', label: 'Unpaid' },
  }
  const STATUS_OPTIONS = [
    { title: 'Unpaid', value: 'unpaid' },
    { title: 'Partial', value: 'partial' },
    { title: 'Paid', value: 'paid' },
  ]

  const { branches, fetchBranches } = useBranches()
  const {
    loanCategories,
    fetchLoanCategories,
    createLoanCategory,
    updateLoanCategory,
    deleteLoanCategory,
  } = useLoanCategories()

  onMounted(() => {
    fetchBranches()
    fetchLoanCategories(true)
  })

  const {
    selected: selectedLoans,
    deleting: loanDeleting,
    error: loanSelError,
    reset: resetLoanSel,
    run: runLoanBulk,
  } = useBulkDelete()

  async function bulkDeleteLoans () {
    try {
      await runLoanBulk('/loans/bulk_delete', { ids: selectedLoans.value })
      resetLoanSel()
      reload()
    } catch { /* error shown in the bar */ }
  }

  const activeCategories = computed(() => loanCategories.value.filter(c => c.is_active))

  function categoryFor (id) {
    return loanCategories.value.find(c => c.id === id) ?? null
  }

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? (id ? `Branch #${id}` : '—')
  }

  function formatMoney (cents) {
    return (Number(cents) / 100).toFixed(2)
  }

  function remaining (loan) {
    return Math.max(0, Number(loan.amount) - Number(loan.paid_amount))
  }

  // Overpayment guard. We can only compare when the payment is in the loan's
  // own currency; cross-currency amounts are left to the API to validate.
  function payAmountError (loan) {
    const pf = payFormFor(loan.id)
    const amount = Number(pf.amount)
    if (!pf.amount || Number.isNaN(amount)) return ''
    if (amount <= 0) return 'Must be greater than 0.'
    const sameCurrency = (pf.currencyCode ?? '').trim() === loan.currency_code
    if (sameCurrency && Math.round(amount * 100) > remaining(loan)) {
      return `Exceeds remaining balance of ${formatMoney(remaining(loan))} ${loan.currency_code}.`
    }
    return ''
  }

  function paidPercent (loan) {
    const total = Number(loan.amount)
    if (!total) return 0
    return Math.min(100, Math.max(0, (Number(loan.paid_amount) / total) * 100))
  }

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
  }

  const headers = computed(() => {
    const base = tab.value === 'branch'
      ? [
        { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
        { title: 'Description', key: 'description', align: 'start', sortable: false },
      ]
      : [
        { title: 'ID', key: 'id', align: 'start', width: 80 },
        { title: 'Description', key: 'description', align: 'start' },
      ]
    return [
      ...base,
      { title: 'Category', key: 'category_id', align: 'start', sortable: false },
      { title: 'Amount', key: 'amount', align: 'end', sortable: false },
      { title: 'Paid', key: 'paid_amount', align: 'end', sortable: false },
      { title: 'Status', key: 'status', align: 'start', sortable: false },
      { title: 'Date', key: 'occurred_at', align: 'start', sortable: false },
      { title: '', key: 'actions', align: 'end', sortable: false, width: 56 },
    ]
  })

  const tab = ref('own')
  const search = ref('')
  const statusFilter = ref(null)
  const selectedCategory = ref(null)
  const branchFilter = ref(null)
  const expanded = ref([])
  const rows = ref([])
  const loadError = ref('')
  const tableRef = ref(null)

  // Per-loan payment lists + add-payment forms, keyed by loan id; populated
  // when a row is expanded, pruned when it leaves the page.
  const payments = reactive({})
  const payForms = reactive({})

  // Server-side filters — ServerSideTable merges these into every request and
  // refetches from page 1 whenever they change. listLoans has no sort_by, so
  // no sortKeys: column headers stay unsorted rather than faking it.
  const queryParams = computed(() => ({
    origin: tab.value === 'branch' ? 'branch_loan' : 'central_loan',
    category_id: selectedCategory.value || '',
    status: statusFilter.value || '',
    branch_id: tab.value === 'branch' ? (branchFilter.value || '') : '',
  }))

  function reload () {
    tableRef.value?.reload()
  }

  function onRowsLoaded (items) {
    rows.value = items
    // Drop cached payment state for rows that fell off the page.
    const ids = new Set(items.map(r => r.id))
    for (const key of Object.keys(payments)) {
      if (!ids.has(Number(key))) delete payments[key]
    }
    for (const key of Object.keys(payForms)) {
      if (!ids.has(Number(key))) delete payForms[key]
    }
  }

  watch(tab, () => {
    resetLoanSel()
    tableRef.value?.clearSelection()
  })

  // -- payments (expanded row) --

  function paymentsFor (loanId) {
    return payments[loanId] ?? { list: [], loading: false, error: '' }
  }

  function payFormFor (loanId) {
    if (!payForms[loanId]) {
      const loan = rows.value.find(r => r.id === loanId)
      payForms[loanId] = {
        amount: null,
        currencyCode: loan?.currency_code ?? '',
        note: '',
        submitting: false,
      }
    }
    return payForms[loanId]
  }

  async function loadPayments (loanId) {
    payments[loanId] = { list: [], loading: true, error: '' }
    try {
      const url = new URL(`${API_BASE}/loan_payments`)
      url.searchParams.set('loan_id', loanId)
      url.searchParams.set('page_size', 100)
      const res = await authFetch(url.toString())
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      const data = await res.json()
      payments[loanId] = { list: data.loan_payments ?? [], loading: false, error: '' }
    } catch (error) {
      payments[loanId] = { list: [], loading: false, error: error.message }
    }
  }

  function onExpandedChange (ids) {
    expanded.value = ids
    for (const id of ids) {
      if (!payments[id]) loadPayments(id)
    }
  }

  async function addPayment (loan) {
    const pf = payFormFor(loan.id)
    if (!(Number(pf.amount) > 0) || (pf.currencyCode ?? '').trim().length < 2) return
    if (payAmountError(loan)) return
    pf.submitting = true
    try {
      const res = await authFetch(`${API_BASE}/loan_payments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          loan_id: loan.id,
          amount: Math.round(Number(pf.amount) * 100),
          currency_code: pf.currencyCode.trim(),
          note: pf.note ?? '',
        }),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      pf.amount = null
      pf.note = ''
      await loadPayments(loan.id)
      reload()
    } catch (error) {
      payments[loan.id] = { ...paymentsFor(loan.id), error: error.message }
    } finally {
      pf.submitting = false
    }
  }

  async function removePayment (loan, payment) {
    if (!confirm(`Delete this payment of ${formatMoney(payment.amount)} ${payment.currency_code}?`)) return
    try {
      const res = await authFetch(`${API_BASE}/loan_payments/${payment.id}`, { method: 'DELETE' })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      await loadPayments(loan.id)
      reload()
    } catch (error) {
      payments[loan.id] = { ...paymentsFor(loan.id), error: error.message }
    }
  }

  // -- add / edit loan --

  const dialog = ref(false)
  const editingId = ref(null)
  const submitting = ref(false)
  const submitError = ref('')
  const form = reactive({ description: '', categoryId: null, amount: null, currencyCode: '' })
  const formErrors = reactive({ description: '', categoryId: '', amount: '', currencyCode: '' })

  function resetForm () {
    Object.assign(form, { description: '', categoryId: null, amount: null, currencyCode: '' })
    Object.assign(formErrors, { description: '', categoryId: '', amount: '', currencyCode: '' })
    submitError.value = ''
  }

  function openCreate () {
    editingId.value = null
    resetForm()
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    resetForm()
    Object.assign(form, {
      description: item.description,
      categoryId: item.category_id,
      amount: Number(item.amount) / 100,
      currencyCode: item.currency_code,
    })
    dialog.value = true
  }

  function closeDialog () {
    dialog.value = false
    editingId.value = null
    resetForm()
  }

  function validate () {
    formErrors.description = form.description?.trim().length >= 2 ? '' : 'Description is required.'
    formErrors.categoryId = form.categoryId ? '' : 'Category is required.'
    formErrors.amount = Number(form.amount) > 0 ? '' : 'Amount must be greater than 0.'
    formErrors.currencyCode = form.currencyCode?.trim().length >= 2 ? '' : 'Currency code is required.'
    return !formErrors.description && !formErrors.categoryId && !formErrors.amount && !formErrors.currencyCode
  }

  async function submit () {
    if (!validate()) return
    submitting.value = true
    submitError.value = ''
    try {
      const body = {
        description: form.description.trim(),
        category_id: form.categoryId,
        amount: Math.round(Number(form.amount) * 100),
        currency_code: form.currencyCode.trim(),
      }
      const res = await authFetch(`${API_BASE}/loans`, {
        method: editingId.value ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(editingId.value ? { id: editingId.value, ...body } : body),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      closeDialog()
      reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  }

  async function removeLoan () {
    if (!editingId.value) return
    if (!confirm('Delete this loan? Loans that still have payments can\'t be deleted.')) return
    submitting.value = true
    submitError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/loans/${editingId.value}`, { method: 'DELETE' })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      closeDialog()
      reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  }

  // -- categories --

  const categoryDialog = ref(false)
  const editingCategory = ref(null)

  function openCreateCategory () {
    editingCategory.value = null
    categoryDialog.value = true
  }

  function openEditCategory (category) {
    editingCategory.value = category
    categoryDialog.value = true
  }

  async function submitCategory (payload) {
    await (editingCategory.value
      ? updateLoanCategory({ id: editingCategory.value.id, ...payload })
      : createLoanCategory(payload))
    await fetchLoanCategories(true)
    reload()
  }

  async function removeCategory (category) {
    if (!confirm(`Delete category "${category.name}"? Categories still used by a loan can't be deleted.`)) return
    try {
      await deleteLoanCategory(category.id)
      if (selectedCategory.value === category.id) selectedCategory.value = null
      await fetchLoanCategories(true)
    } catch (error) {
      loadError.value = error.message
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
.loans-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.loans-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.loans-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.table-search {
  flex: 0 1 320px;
}
.loans-expanded td {
  background: rgba(var(--v-theme-primary), 0.03);
}
</style>
