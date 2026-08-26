<template>
  <v-dialog v-model="dialog" max-width="480">
    <v-card class="px-4">
      <v-card-title>{{ editingId ? 'Edit Invoice Type' : 'Add a New Invoice Type' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
          />
          <v-text-field
            v-model="code.value.value"
            density="compact"
            :error-messages="code.errorMessage.value"
            hint="Short code used in invoice numbers, e.g. RT, WS"
            label="Code"
            persistent-hint
          />
          <v-switch
            v-model="isActive.value.value"
            color="primary"
            density="compact"
            hide-details
            label="Active"
          />

          <v-alert v-if="submitError" class="mb-4 mt-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingId ? 'Save' : 'Add Invoice Type'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Invoice Types</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Invoice Type" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="invoice_types"
  >
    <template #item.is_default="{ item }">
      <v-chip v-if="item.is_default" color="primary" size="small">Default</v-chip>
    </template>
    <template #item.is_active="{ item }">
      <v-chip :color="item.is_active ? 'success' : 'default'" size="small">
        {{ item.is_active ? 'Active' : 'Inactive' }}
      </v-chip>
    </template>
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEdit(item)" />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInvoiceTypes } from '@/composables/useInvoiceTypes'
  import { API_BASE } from '@/config'

  const { createInvoiceType, updateInvoiceType } = useInvoiceTypes()

  const apiURL = `${API_BASE}/invoice_types`
  const headers = ref([
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: '', key: 'is_default', align: 'center', sortable: false },
    { title: 'Status', key: 'is_active', align: 'center', sortable: false },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const editingId = ref(null)

  const { handleSubmit, handleReset, setValues } = useForm({
    validationSchema: {
      name (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
      code (value) {
        if (value?.length >= 1) return true
        return 'Code is required.'
      },
    },
  })

  const name = useField('name')
  const code = useField('code')
  const isActive = useField('isActive')

  function openCreate () {
    editingId.value = null
    handleReset()
    isActive.value.value = true
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    setValues({
      name: item.name,
      code: item.code,
      isActive: item.is_active,
    })
    dialog.value = true
  }

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
    editingId.value = null
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      await (editingId.value
        ? updateInvoiceType({
          id: editingId.value,
          name: values.name,
          code: values.code,
          is_active: !!values.isActive,
        })
        : createInvoiceType({
          name: values.name,
          code: values.code,
          is_active: !!values.isActive,
          is_default: false,
        }))
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })
</script>
