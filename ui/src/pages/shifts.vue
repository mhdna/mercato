<template>
  <v-dialog v-model="dialog" max-width="420" :persistent="true">
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

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Shifts</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Open Shift" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="shifts"
  >
    <template #item.is_closed="{ item }">
      <v-chip :color="item.is_closed ? 'default' : 'success'" size="small">
        {{ item.is_closed ? 'Closed' : 'Open' }}
      </v-chip>
    </template>
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
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useCashboxes } from '@/composables/useCashboxes'
  import { useShifts } from '@/composables/useShifts'
  import { API_BASE } from '@/config'

  const { createShift, closeShift } = useShifts()
  const { cashboxes, fetchCashboxes } = useCashboxes()
  fetchCashboxes()

  const apiURL = `${API_BASE}/shifts`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Cashbox ID', key: 'cashbox_id', align: 'start' },
    { title: 'Status', key: 'is_closed', align: 'start' },
    { title: 'Opened At', key: 'created_at', align: 'end' },
    { title: 'Closed At', key: 'closed_at', align: 'end' },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const closingId = ref(null)

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
    } finally {
      closingId.value = null
    }
  }
</script>
