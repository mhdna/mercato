<template>
  <v-dialog v-model="dialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>Add Loan Payment</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-select
            v-model="loanId.value.value"
            density="compact"
            :error-messages="loanId.errorMessage.value"
            item-title="description"
            item-value="id"
            :items="loans"
            label="Loan"
          >
            <template #item="{ props: itemProps, item }">
              <v-list-item v-bind="itemProps" :subtitle="`${formatMoney(item.raw.amount)} ${item.raw.currency_code}`" />
            </template>
          </v-select>
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
          <v-text-field
            v-model="note.value.value"
            density="compact"
            label="Note"
          />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" text="Add Payment" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-container>
    <v-card>
      <div class="d-flex justify-space-between align-center pa-4">
        <h2 class="text-h6">Loan Payments</h2>
        <v-select
          v-model="loanFilter"
          clearable
          density="compact"
          hide-details
          item-title="description"
          item-value="id"
          :items="loans"
          placeholder="All loans"
          style="max-width: 260px"
          variant="outlined"
          @update:model-value="onLoanFilterChange"
        />
        <v-btn color="primary" prepend-icon="mdi-plus" text="Add Payment" @click="openCreate" />
      </div>

      <v-card-text>
        <ServerSideTable
          ref="tableRef"
          :api-u-r-l="buildUrl()"
          :headers="headers"
          root-key="loan_payments"
        >
          <template #item.loan_id="{ item }">
            {{ loanDescription(item.loan_id) }}
          </template>
          <template #item.amount="{ item }">
            {{ formatMoney(item.amount) }} {{ item.currency_code }}
          </template>
          <template #item.paid_at="{ item }">
            {{ new Date(item.paid_at).toLocaleString() }}
          </template>
          <template #item.actions="{ item }">
            <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="removePayment(item)" />
          </template>
        </ServerSideTable>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { onMounted, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { authFetch } from '@/composables/useApi'
  import { useLoans } from '@/composables/useLoans'
  import { API_BASE } from '@/config'

  const { loans, fetchLoans } = useLoans()
  onMounted(() => fetchLoans())

  function loanDescription (id) {
    return loans.value.find(l => l.id === id)?.description ?? `Loan #${id}`
  }

  function formatMoney (cents) {
    return (Number(cents) / 100).toFixed(2)
  }

  const headers = [
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Loan', key: 'loan_id', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Note', key: 'note', align: 'start', sortable: false },
    { title: 'Paid At', key: 'paid_at', align: 'start' },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ]

  const tableRef = ref(null)
  const loanFilter = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  function buildUrl () {
    const url = new URL(`${API_BASE}/loan_payments`, window.location.origin)
    if (loanFilter.value) url.searchParams.set('loan_id', loanFilter.value)
    return url.toString()
  }

  function onLoanFilterChange () {
    tableRef.value?.reload()
  }

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      loanId (value) {
        if (value) return true
        return 'Loan is required.'
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

  const loanId = useField('loanId')
  const amount = useField('amount')
  const currencyCode = useField('currencyCode')
  const note = useField('note')

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
      const res = await authFetch(`${API_BASE}/loan_payments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          loan_id: values.loanId,
          amount: Math.round(Number(values.amount) * 100),
          currency_code: values.currencyCode,
          note: values.note ?? '',
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

  async function removePayment (item) {
    if (!confirm(`Delete this payment of ${formatMoney(item.amount)} ${item.currency_code}?`)) return
    try {
      const res = await authFetch(`${API_BASE}/loan_payments/${item.id}`, { method: 'DELETE' })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    }
  }
</script>
