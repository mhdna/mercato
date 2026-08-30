<template>
  <v-dialog
    max-width="760"
    :model-value="modelValue"
    scrollable
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <v-card rounded="lg">
      <v-card-title class="d-flex align-center ga-2 px-4 py-3">
        <v-icon icon="mdi-tag-multiple" />
        <span>Invoice Types</span>
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Invoice Type"
          variant="flat"
          @click="openCreate"
        />
        <v-btn icon="mdi-close" size="small" variant="text" @click="closeManager" />
      </v-card-title>
      <v-divider />

      <v-card-text class="pa-4">
        <v-alert
          v-if="loadError"
          class="mb-3"
          density="compact"
          type="error"
          variant="tonal"
        >
          {{ loadError }}
        </v-alert>
        <v-data-table
          class="invoice-types-table"
          density="comfortable"
          :headers="headers"
          :items="invoiceTypes"
          :items-per-page-options="[14, 25, 50, 100]"
          :loading="loading"
          @click:row="openRow"
        >
          <template #item.is_default="{ item }">
            <v-chip v-if="item.is_default" color="primary" size="small">Default</v-chip>
          </template>
          <template #item.is_active="{ item }">
            <v-chip :color="item.is_active ? 'success' : 'default'" size="small">
              {{ item.is_active ? 'Active' : 'Inactive' }}
            </v-chip>
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="formDialog" max-width="420">
    <v-card rounded="lg">
      <v-card-title class="d-flex align-center px-4 py-3 text-h6">
        {{ editingId ? 'Edit Invoice Type' : 'Add Invoice Type' }}
        <v-spacer />
        <v-btn icon="mdi-close" size="small" variant="text" @click="closeForm" />
      </v-card-title>
      <v-divider />
      <v-card-text class="px-4 pt-4">
        <form @submit.prevent="submit">
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
            variant="outlined"
          />
          <v-text-field
            v-model="code.value.value"
            density="compact"
            :error-messages="code.errorMessage.value"
            hint="Short code used in invoice numbers, e.g. RT, WS"
            label="Code"
            persistent-hint
            variant="outlined"
          />
          <v-switch
            v-model="isActive.value.value"
            color="primary"
            density="compact"
            hide-details
            label="Active"
          />

          <v-alert
            v-if="submitError"
            class="mt-4"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0 pt-4">
            <v-spacer />
            <v-btn text="Cancel" variant="text" @click="closeForm" />
            <v-btn
              color="primary"
              :loading="submitting"
              :text="editingId ? 'Save' : 'Add'"
              type="submit"
              variant="flat"
            />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { ref, watch } from 'vue'
  import { useInvoiceTypes } from '@/composables/useInvoiceTypes'

  const props = defineProps({
    modelValue: {
      type: Boolean,
      default: false,
    },
  })

  const emit = defineEmits(['update:modelValue'])
  const { invoiceTypes, fetchInvoiceTypes, createInvoiceType, updateInvoiceType } = useInvoiceTypes()
  const loading = ref(false)
  const loadError = ref('')
  const formDialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const editingId = ref(null)

  const headers = [
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: '', key: 'is_default', align: 'center', sortable: false },
    { title: 'Status', key: 'is_active', align: 'center', sortable: false },
  ]

  const { handleSubmit, handleReset, setValues } = useForm({
    validationSchema: {
      name (value) {
        return value?.length >= 2 || 'Name needs to be at least 2 characters.'
      },
      code (value) {
        return value?.length >= 1 || 'Code is required.'
      },
    },
  })

  const name = useField('name')
  const code = useField('code')
  const isActive = useField('isActive')

  async function loadInvoiceTypes () {
    loading.value = true
    loadError.value = ''
    try {
      await fetchInvoiceTypes(true)
    } catch (error) {
      loadError.value = error.message
    } finally {
      loading.value = false
    }
  }

  watch(() => props.modelValue, open => {
    if (open) loadInvoiceTypes()
  })

  function openCreate () {
    editingId.value = null
    handleReset()
    isActive.value.value = true
    submitError.value = ''
    formDialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    setValues({ name: item.name, code: item.code, isActive: item.is_active })
    submitError.value = ''
    formDialog.value = true
  }

  function openRow (_event, { item }) {
    openEdit(item)
  }

  function closeForm () {
    formDialog.value = false
    handleReset()
    submitError.value = ''
    editingId.value = null
  }

  function closeManager () {
    emit('update:modelValue', false)
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
      closeForm()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })
</script>

<style scoped>
:deep(.invoice-types-table tbody tr) {
  cursor: pointer;
}
</style>
