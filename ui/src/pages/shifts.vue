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

    <v-card class="shifts-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-clock-outline" />
        <span>Shifts</span>
        <v-spacer />
        <v-text-field
          v-if="tab === 'central'"
          v-model="search"
          class="shifts-search"
          clearable
          density="compact"
          hide-details
          label="Search cashbox or #id"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-select
          v-if="tab === 'central'"
          v-model="statusFilter"
          clearable
          density="compact"
          hide-details
          :items="STATUS_OPTIONS"
          placeholder="Any status"
          style="max-width: 150px"
          variant="outlined"
        />
        <v-select
          v-if="tab === 'central'"
          v-model="cashboxFilter"
          clearable
          density="compact"
          hide-details
          :item-title="c => `${c.name} (${c.code})`"
          :item-value="c => c.id"
          :items="cashboxes"
          placeholder="Any cashbox"
          style="max-width: 220px"
          variant="outlined"
        />
        <v-btn-toggle
          v-model="tab"
          color="primary"
          density="comfortable"
          mandatory
          variant="outlined"
        >
          <v-btn value="central">Central</v-btn>
          <v-btn value="branch">Branch</v-btn>
        </v-btn-toggle>
        <v-btn
          v-if="tab === 'central'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Open Shift"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <!-- ------------------------------------------------------ CENTRAL -->
      <div v-if="tab === 'central'" class="shifts-content">
        <ServerSideTable
          ref="tableRef"
          :api-u-r-l="`${API_BASE}/shifts`"
          class="shifts-table"
          density="comfortable"
          :external-search="search ?? ''"
          flush
          :headers="centralHeaders"
          :query-params="centralQueryParams"
          root-key="shifts"
          :show-search-icon="false"
          :sort-keys="['id', 'cashbox', 'created_at', 'closed_at']"
        >
          <template #item.cashbox="{ item }">
            <div class="d-flex flex-column">
              <span>{{ item.cashbox_name }}</span>
              <span class="text-medium-emphasis text-caption">{{ item.cashbox_code }}</span>
            </div>
          </template>
          <template #item.is_closed="{ item }">
            <v-chip :color="item.is_closed ? 'default' : 'success'" size="small" variant="tonal">
              {{ item.is_closed ? 'Closed' : 'Open' }}
            </v-chip>
          </template>
          <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
          <template #item.closed_at="{ item }">{{ formatDate(item.closed_at) }}</template>
          <template #item.actions="{ item }">
            <v-btn
              :disabled="item.is_closed"
              :loading="closingId === item.id"
              size="small"
              text="Close"
              variant="text"
              @click="close(item)"
            />
          </template>
        </ServerSideTable>
      </div>

      <!-- ------------------------------------------------------- BRANCH -->
      <div v-else class="shifts-layout">
        <PageSidebar
          v-model="selectedBranch"
          all-title="All branches"
          empty-text="No branches yet."
          :items="branchItems"
          :loading="branchesLoading"
          show-all
        />

        <v-divider vertical />

        <div class="shifts-content">
          <ServerSideTable
            :api-u-r-l="`${API_BASE}/branch_shifts`"
            class="shifts-table"
            density="comfortable"
            flush
            :headers="branchHeaders"
            :query-params="branchQueryParams"
            root-key="branch_shifts"
            :show-search-icon="false"
          >
            <template #item.branch_id="{ item }">{{ branchName(item.branch_id) }}</template>
            <template #item.opened_at="{ item }">{{ formatDate(item.opened_at) }}</template>
            <template #item.closed_at="{ item }">{{ formatDate(item.closed_at) }}</template>
            <template #item.expected_usd="{ item }">{{ money(item.expected_usd) }}</template>
            <template #item.counted_usd="{ item }">{{ money(item.counted_usd) }}</template>
            <template #item.variance_usd="{ item }">
              <v-chip :color="varianceColor(item.variance_usd)" size="small" variant="tonal">
                {{ signedMoney(item.variance_usd) }}
              </v-chip>
            </template>
            <template #item.received_at="{ item }">{{ formatDate(item.received_at) }}</template>
          </ServerSideTable>
        </div>
      </div>
    </v-card>

    <v-dialog v-model="dialog" max-width="420">
      <v-card class="px-4">
        <v-card-title>Open a New Shift</v-card-title>
        <v-card-text>
          <form @submit.prevent="submit">
            <v-select
              v-model="cashboxId.value.value"
              density="compact"
              :error-messages="cashboxId.errorMessage.value"
              :item-title="c => `${c.name} (${c.code})`"
              :item-value="c => c.id"
              :items="cashboxes"
              label="Cashbox"
            />

            <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
              {{ submitError }}
            </v-alert>

            <v-card-actions class="px-0">
              <v-spacer />
              <v-btn text="Cancel" @click="closeDialog" />
              <v-btn color="primary" :loading="submitting" text="Open Shift" type="submit" />
            </v-card-actions>
          </form>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { computed, ref } from 'vue'
  import PageSidebar from '@/components/PageSidebar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBranches } from '@/composables/useBranches'
  import { useCashboxes } from '@/composables/useCashboxes'
  import { useShifts } from '@/composables/useShifts'
  import { API_BASE } from '@/config'

  const STATUS_OPTIONS = [
    { title: 'Open', value: 'open' },
    { title: 'Closed', value: 'closed' },
  ]

  const { createShift, closeShift } = useShifts()
  const { cashboxes, fetchCashboxes } = useCashboxes()
  const { branches, fetchBranches } = useBranches()
  fetchCashboxes()

  const branchesLoading = ref(true)
  fetchBranches().finally(() => { branchesLoading.value = false })

  const tab = ref('central')

  const centralHeaders = [
    { title: 'ID', key: 'id', align: 'start', width: 80 },
    { title: 'Cashbox', key: 'cashbox', align: 'start' },
    { title: 'Status', key: 'is_closed', align: 'start' },
    { title: 'Opened At', key: 'created_at', align: 'end' },
    { title: 'Closed At', key: 'closed_at', align: 'end' },
    { title: '', key: 'actions', align: 'end', sortable: false, width: 90 },
  ]

  const branchHeaders = [
    { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
    { title: 'Closed By', key: 'closing_person_name', align: 'start', sortable: false },
    { title: 'Opened', key: 'opened_at', align: 'start', sortable: false },
    { title: 'Closed', key: 'closed_at', align: 'start', sortable: false },
    { title: 'Expected USD', key: 'expected_usd', align: 'end', sortable: false },
    { title: 'Counted USD', key: 'counted_usd', align: 'end', sortable: false },
    { title: 'Variance USD', key: 'variance_usd', align: 'end', sortable: false },
    { title: 'Received', key: 'received_at', align: 'end', sortable: false },
  ]

  const search = ref('')
  const statusFilter = ref(null)
  const cashboxFilter = ref(null)
  const selectedBranch = ref(null)

  const branchItems = computed(() => branches.value.map(b => ({
    value: b.id,
    title: b.name,
    subtitle: b.code,
    raw: b,
  })))

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? (id ? `Branch #${id}` : '—')
  }

  // ServerSideTable merges these into every request and refetches from page 1
  // whenever they change (see the API's listShifts / listBranchShifts params).
  const centralQueryParams = computed(() => ({
    status: statusFilter.value || '',
    cashbox_id: cashboxFilter.value || '',
  }))
  const branchQueryParams = computed(() => ({
    branch_id: selectedBranch.value || '',
  }))

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const closingId = ref(null)
  const loadError = ref('')

  function money (cents) {
    return (Number(cents) / 100).toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  }

  function signedMoney (cents) {
    const value = money(cents)
    return Number(cents) > 0 ? `+${value}` : value
  }

  function varianceColor (cents) {
    if (Number(cents) < 0) return 'error'
    if (Number(cents) > 0) return 'warning'
    return 'default'
  }

  function formatDate (value) {
    // closed_at / opened_at come back as Go's sql.NullTime -> { Time, Valid }.
    if (value && typeof value === 'object') {
      if (!value.Valid) return '—'
      value = value.Time
    }
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
  }

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      cashboxId (value) {
        return !!value || 'Select a cashbox.'
      },
    },
  })

  const cashboxId = useField('cashboxId')

  function openCreate () {
    handleReset()
    submitError.value = ''
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
      await createShift({ cashbox_id: values.cashboxId })
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  async function close (item) {
    closingId.value = item.id
    try {
      await closeShift(item.id)
      tableRef.value?.reload()
    } catch (error) {
      loadError.value = error.message
    } finally {
      closingId.value = null
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
.shifts-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.shifts-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.shifts-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.shifts-search {
  flex: 0 1 260px;
}
</style>
