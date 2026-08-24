<template>
  <v-dialog v-model="dialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>{{ editingCode ? 'Edit Currency' : 'Add a New Currency' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="code.value.value"
            density="compact"
            :disabled="!!editingCode"
            :error-messages="code.errorMessage.value"
            hint="e.g. USD, EUR"
            label="Code"
          />
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
          />
          <v-text-field
            v-model="symbol.value.value"
            density="compact"
            :error-messages="symbol.errorMessage.value"
            label="Symbol"
          />
          <v-text-field
            v-model.number="valueInDefaultCurrency.value.value"
            density="compact"
            :error-messages="valueInDefaultCurrency.errorMessage.value"
            hint="How many units of the default currency 1 unit of this currency is worth"
            label="Value in Default Currency"
            persistent-hint
            type="number"
          />

          <v-alert v-if="submitError" class="mb-4 mt-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingCode ? 'Save' : 'Add Currency'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="deleteDialog" max-width="420">
    <v-card>
      <v-card-title>Delete Currency</v-card-title>
      <v-card-text>Are you sure you want to delete "{{ deleteTarget?.name }}"? This cannot be undone.</v-card-text>
      <v-alert v-if="deleteError" class="mx-4 mb-2" type="error" variant="tonal">{{ deleteError }}</v-alert>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" @click="deleteDialog = false" />
        <v-btn color="error" :loading="deleting" text="Delete" @click="confirmDelete" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Currencies</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Currency" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="currencies"
  >
    <template #item.is_default="{ item }">
      <v-chip v-if="item.is_default" color="primary" size="small">Default</v-chip>
    </template>
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEdit(item)" />
      <v-icon-btn
        color="error"
        :disabled="item.is_default"
        icon="mdi-delete"
        size="small"
        variant="text"
        @click="openDelete(item)"
      />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useCurrencies } from '@/composables/useCurrencies'
  import { API_BASE } from '@/config'

  const { createCurrency, updateCurrency, deleteCurrency } = useCurrencies()

  const apiURL = `${API_BASE}/currencies/`
  const headers = ref([
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Symbol', key: 'symbol', align: 'start' },
    { title: 'Value in Default Currency', key: 'value_in_default_currency', align: 'end' },
    { title: '', key: 'is_default', align: 'center', sortable: false },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')
  const editingCode = ref(null)

  const { handleSubmit, handleReset, setValues } = useForm({
    validationSchema: {
      code (value) {
        if (editingCode.value) return true
        return (value?.length >= 2) || 'Code is required.'
      },
      name (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
      symbol (value) {
        if (value?.length >= 1) return true
        return 'Symbol is required.'
      },
      valueInDefaultCurrency (value) {
        if (Number(value) > 0) return true
        return 'Must be greater than 0.'
      },
    },
  })

  const code = useField('code')
  const name = useField('name')
  const symbol = useField('symbol')
  const valueInDefaultCurrency = useField('valueInDefaultCurrency')

  function openCreate () {
    editingCode.value = null
    handleReset()
    valueInDefaultCurrency.value.value = 1
    dialog.value = true
  }

  function openEdit (item) {
    editingCode.value = item.code
    setValues({
      code: item.code,
      name: item.name,
      symbol: item.symbol,
      valueInDefaultCurrency: item.value_in_default_currency,
    })
    dialog.value = true
  }

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
    editingCode.value = null
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const payload = {
        code: values.code,
        name: values.name,
        symbol: values.symbol,
        value_in_default_currency: values.valueInDefaultCurrency,
      }
      await (editingCode.value ? updateCurrency(payload) : createCurrency(payload))
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  const deleteDialog = ref(false)
  const deleteTarget = ref(null)
  const deleting = ref(false)
  const deleteError = ref('')

  function openDelete (item) {
    deleteTarget.value = item
    deleteError.value = ''
    deleteDialog.value = true
  }

  async function confirmDelete () {
    deleting.value = true
    deleteError.value = ''
    try {
      await deleteCurrency(deleteTarget.value.code)
      deleteDialog.value = false
      tableRef.value?.reload()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }
</script>
