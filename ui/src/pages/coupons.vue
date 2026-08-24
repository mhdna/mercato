<template>
  <v-dialog v-model="dialog" max-width="560" :persistent="true">
    <v-card class="px-4">
      <v-card-title>Add a New Coupon</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="code.value.value"
            density="compact"
            :error-messages="code.errorMessage.value"
            label="Code"
          />
          <v-select
            v-model="clientId.value.value"
            density="compact"
            :error-messages="clientId.errorMessage.value"
            :item-title="c => `${c.name} (${c.phone})`"
            :item-value="c => c.id"
            :items="clients"
            label="Client"
          />
          <v-row>
            <v-col cols="6">
              <v-select
                v-model="discountType.value.value"
                density="compact"
                :error-messages="discountType.errorMessage.value"
                :items="['fixed', 'percentage']"
                label="Discount Type"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model="validUntil.value.value"
                density="compact"
                :error-messages="validUntil.errorMessage.value"
                label="Valid Until"
                type="date"
              />
            </v-col>
          </v-row>
          <v-textarea
            v-model="reason.value.value"
            density="compact"
            :error-messages="reason.errorMessage.value"
            label="Reason"
            no-resize
            rows="2"
          />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" text="Add Coupon" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Coupons</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Coupon" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="coupons"
  >
    <template #item.status="{ item }">
      <v-chip :color="item.status === 'active' ? 'success' : 'default'" size="small">{{ item.status }}</v-chip>
    </template>
    <template #item.actions="{ item }">
      <v-btn
        :disabled="item.status !== 'active'"
        size="small"
        text="Deactivate"
        variant="text"
        @click="deactivate(item)"
      />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useClients } from '@/composables/useClients'
  import { useCoupons } from '@/composables/useCoupons'
  import { API_BASE } from '@/config'

  const { createCoupon, deactivateCoupon } = useCoupons()
  const { clients, fetchClients } = useClients()
  fetchClients()

  const apiURL = `${API_BASE}/coupons`
  const headers = ref([
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Status', key: 'status', align: 'start' },
    { title: 'Discount Type', key: 'discount_type', align: 'start' },
    { title: 'Reason', key: 'reason', align: 'start' },
    { title: 'Client ID', key: 'client_id', align: 'end' },
    { title: 'Valid Until', key: 'valid_until', align: 'end' },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      code (value) {
        if (value?.length >= 1) return true
        return 'Code is required.'
      },
      clientId () {
        return true
      },
      discountType (value) {
        return ['fixed', 'percentage'].includes(value) || 'Select a discount type.'
      },
      validUntil (value) {
        return !!value || 'Valid until date is required.'
      },
      reason (value) {
        if (value?.length >= 1) return true
        return 'Reason is required.'
      },
    },
  })

  const code = useField('code')
  const clientId = useField('clientId')
  const discountType = useField('discountType')
  const validUntil = useField('validUntil')
  const reason = useField('reason')

  function openCreate () {
    handleReset()
    discountType.value.value = 'percentage'
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
      await createCoupon({
        code: values.code,
        status: 'active',
        discount_type: values.discountType,
        reason: values.reason,
        client_id: values.clientId,
        valid_until: new Date(values.validUntil).toISOString(),
      })
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  async function deactivate (item) {
    try {
      await deactivateCoupon(item.code)
      tableRef.value?.reload()
    } catch {
      // surfaced via table reload/error alert on next fetch
    }
  }
</script>
