<template>
  <v-dialog v-model="dialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>Add Expense</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="description.value.value"
            density="compact"
            :error-messages="description.errorMessage.value"
            label="Description"
          />
          <v-text-field
            v-model="category.value.value"
            density="compact"
            :error-messages="category.errorMessage.value"
            label="Category"
          />
          <v-text-field
            v-model.number="amount.value.value"
            density="compact"
            :error-messages="amount.errorMessage.value"
            label="Amount"
            type="number"
          />
          <v-text-field
            v-model="currencyCode.value.value"
            density="compact"
            :error-messages="currencyCode.errorMessage.value"
            label="Currency Code"
          />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" text="Add Expense" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="recurringDialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>Add Recurring Expense</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitRecurring">
          <v-text-field
            v-model="rDescription.value.value"
            density="compact"
            :error-messages="rDescription.errorMessage.value"
            label="Description"
          />
          <v-text-field
            v-model="rCategory.value.value"
            density="compact"
            :error-messages="rCategory.errorMessage.value"
            label="Category"
          />
          <v-text-field
            v-model.number="rAmount.value.value"
            density="compact"
            :error-messages="rAmount.errorMessage.value"
            label="Amount"
            type="number"
          />
          <v-text-field
            v-model="rCurrencyCode.value.value"
            density="compact"
            :error-messages="rCurrencyCode.errorMessage.value"
            label="Currency Code"
          />
          <v-row dense>
            <v-col cols="6">
              <v-text-field
                v-model.number="rIntervalCount.value.value"
                density="compact"
                :error-messages="rIntervalCount.errorMessage.value"
                label="Every"
                type="number"
              />
            </v-col>
            <v-col cols="6">
              <v-select
                v-model="rIntervalUnit.value.value"
                density="compact"
                :error-messages="rIntervalUnit.errorMessage.value"
                :items="[{ title: 'Days', value: 'day' }, { title: 'Months', value: 'month' }]"
                label="Unit"
              />
            </v-col>
          </v-row>
          <v-text-field
            v-model="rNextDueAt.value.value"
            density="compact"
            :error-messages="rNextDueAt.errorMessage.value"
            label="First Due Date"
            type="date"
          />

          <v-alert v-if="recurringSubmitError" class="mb-4" type="error" variant="tonal">
            {{ recurringSubmitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeRecurringDialog" />
            <v-btn color="primary" :loading="recurringSubmitting" text="Add Recurring Expense" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-container>
    <v-card>
      <div class="d-flex justify-space-between align-center pe-4">
        <v-tabs v-model="tab" color="primary">
          <v-tab value="own">Expenses</v-tab>
          <v-tab value="recurring">Recurring</v-tab>
          <v-tab value="branch">Branch Expenses</v-tab>
        </v-tabs>
        <v-btn
          v-if="tab === 'own'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Expense"
          @click="dialog = true"
        />
        <v-btn
          v-else-if="tab === 'recurring'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Recurring Expense"
          @click="recurringDialog = true"
        />
        <v-select
          v-else
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
          @update:model-value="onBranchFilterChange"
        />
      </div>

      <v-window v-model="tab">
        <v-window-item value="own">
          <v-card-text>
            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="`${API_BASE}/expenses`"
              :headers="headers"
              root-key="expenses"
            />
          </v-card-text>
        </v-window-item>

        <v-window-item value="recurring">
          <v-card-text>
            <v-alert v-if="recurringLoadError" class="mb-2" type="error" variant="tonal">
              {{ recurringLoadError }}
            </v-alert>
            <v-data-table density="compact" :headers="recurringHeaders" :items="recurringExpenses" :loading="recurringLoading">
              <template #item.amount="{ item }">
                {{ formatMoney(item.amount) }} {{ item.currency_code }}
              </template>
              <template #item.interval="{ item }">
                Every {{ item.interval_count }} {{ item.interval_unit }}{{ item.interval_count > 1 ? 's' : '' }}
              </template>
              <template #item.next_due_at="{ item }">
                {{ new Date(item.next_due_at).toLocaleDateString() }}
              </template>
              <template #item.active="{ item }">
                <v-switch
                  :model-value="item.active"
                  color="primary"
                  density="compact"
                  hide-details
                  @update:model-value="value => toggleRecurringActive(item, value)"
                />
              </template>
            </v-data-table>
          </v-card-text>
        </v-window-item>

        <v-window-item value="branch">
          <v-card-text>
            <ServerSideTable
              ref="branchTableRef"
              :api-u-r-l="buildBranchExpensesUrl()"
              :headers="branchHeaders"
              root-key="branch_expenses"
            >
              <template #item.branch_id="{ item }">
                {{ branchName(item.branch_id) }}
              </template>
              <template #item.amount="{ item }">
                {{ formatMoney(item.amount) }} {{ item.currency_code }}
              </template>
              <template #item.occurred_at="{ item }">
                {{ new Date(item.occurred_at).toLocaleString() }}
              </template>
              <template #item.received_at="{ item }">
                {{ new Date(item.received_at).toLocaleString() }}
              </template>
            </ServerSideTable>
          </v-card-text>
        </v-window-item>
      </v-window>
    </v-card>
  </v-container>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { onMounted, ref, watch } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBranches } from '@/composables/useBranches'
  import { authFetch } from '@/composables/useApi'
  import { API_BASE } from '@/config'

  const { branches, fetchBranches } = useBranches()
  onMounted(() => fetchBranches())

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  function formatMoney (cents) {
    return (Number(cents) / 100).toFixed(2)
  }

  const tab = ref('own')

  const headers = [
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Category', key: 'category', align: 'start' },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Date', key: 'created_at', align: 'start' },
  ]

  const branchHeaders = [
    { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
    { title: 'Description', key: 'description', align: 'start', sortable: false },
    { title: 'Category', key: 'category', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end', sortable: false },
    { title: 'Occurred At', key: 'occurred_at', align: 'start', sortable: false },
    { title: 'Received At', key: 'received_at', align: 'start', sortable: false },
  ]

  const recurringHeaders = [
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Category', key: 'category', align: 'start' },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Interval', key: 'interval', align: 'start', sortable: false },
    { title: 'Next Due', key: 'next_due_at', align: 'start' },
    { title: 'Active', key: 'active', align: 'center', sortable: false },
  ]

  const tableRef = ref(null)
  const branchTableRef = ref(null)
  const branchFilter = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  function buildBranchExpensesUrl () {
    const url = new URL(`${API_BASE}/branch_expenses`, window.location.origin)
    if (branchFilter.value) url.searchParams.set('branch_id', branchFilter.value)
    return url.toString()
  }

  function onBranchFilterChange () {
    branchTableRef.value?.reload()
  }

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      description (value) {
        if (value?.length >= 2) return true
        return 'Description is required.'
      },
      amount (value) {
        if (value !== undefined && value !== null && value !== '' && Number(value) > 0) return true
        return 'Amount must be greater than 0.'
      },
      currencyCode (value) {
        if (value?.length >= 2) return true
        return 'Currency code is required.'
      },
    },
  })

  const description = useField('description')
  const category = useField('category')
  const amount = useField('amount')
  const currencyCode = useField('currencyCode')

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/expenses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          description: values.description,
          category: values.category,
          amount: Math.round(Number(values.amount) * 100),
          currency_code: values.currencyCode,
        }),
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }

      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  // -- Recurring expenses --

  const recurringExpenses = ref([])
  const recurringLoading = ref(false)
  const recurringLoadError = ref('')

  async function fetchRecurringExpenses () {
    recurringLoading.value = true
    recurringLoadError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/recurring_expenses`)
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      recurringExpenses.value = await res.json()
    } catch (error) {
      recurringLoadError.value = error.message
    } finally {
      recurringLoading.value = false
    }
  }

  watch(tab, value => {
    if (value === 'recurring' && recurringExpenses.value.length === 0) fetchRecurringExpenses()
  })

  async function toggleRecurringActive (item, active) {
    try {
      const res = await authFetch(`${API_BASE}/recurring_expenses/active`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: item.id, active }),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      item.active = active
    } catch (error) {
      recurringLoadError.value = error.message
    }
  }

  const recurringDialog = ref(false)
  const recurringSubmitting = ref(false)
  const recurringSubmitError = ref('')

  const { handleSubmit: handleRecurringSubmit, handleReset: handleRecurringReset } = useForm({
    validationSchema: {
      rDescription (value) {
        if (value?.length >= 2) return true
        return 'Description is required.'
      },
      rAmount (value) {
        if (value !== undefined && value !== null && value !== '' && Number(value) > 0) return true
        return 'Amount must be greater than 0.'
      },
      rCurrencyCode (value) {
        if (value?.length >= 2) return true
        return 'Currency code is required.'
      },
      rIntervalCount (value) {
        if (value !== undefined && value !== null && value !== '' && Number(value) > 0) return true
        return 'Must be greater than 0.'
      },
      rIntervalUnit (value) {
        if (value === 'day' || value === 'month') return true
        return 'Pick an interval unit.'
      },
      rNextDueAt (value) {
        if (value) return true
        return 'First due date is required.'
      },
    },
  })

  const rDescription = useField('rDescription')
  const rCategory = useField('rCategory')
  const rAmount = useField('rAmount')
  const rCurrencyCode = useField('rCurrencyCode')
  const rIntervalCount = useField('rIntervalCount', undefined, { initialValue: 1 })
  const rIntervalUnit = useField('rIntervalUnit', undefined, { initialValue: 'month' })
  const rNextDueAt = useField('rNextDueAt')

  function closeRecurringDialog () {
    recurringDialog.value = false
    handleRecurringReset()
    recurringSubmitError.value = ''
  }

  const submitRecurring = handleRecurringSubmit(async values => {
    recurringSubmitting.value = true
    recurringSubmitError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/recurring_expenses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          description: values.rDescription,
          category: values.rCategory,
          amount: Math.round(Number(values.rAmount) * 100),
          currency_code: values.rCurrencyCode,
          interval_count: Number(values.rIntervalCount),
          interval_unit: values.rIntervalUnit,
          next_due_at: new Date(values.rNextDueAt).toISOString(),
        }),
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }

      closeRecurringDialog()
      fetchRecurringExpenses()
    } catch (error) {
      recurringSubmitError.value = error.message
    } finally {
      recurringSubmitting.value = false
    }
  })
</script>
