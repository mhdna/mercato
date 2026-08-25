<template>
  <v-dialog v-model="dialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>{{ editingId ? 'Edit Client' : 'Add a New Client' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
          />
          <v-text-field
            v-model="phone.value.value"
            density="compact"
            :error-messages="phone.errorMessage.value"
            label="Phone"
          />
          <v-select
            v-model="clientType"
            density="compact"
            :items="[{ title: 'Retail', value: 'retail' }, { title: 'Wholesale', value: 'wholesale' }]"
            label="Type"
          />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingId ? 'Save' : 'Add Client'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="deleteDialog" max-width="420">
    <v-card>
      <v-card-title>Delete Client</v-card-title>
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
    <h2 class="text-h6">Clients</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Client" @click="openCreate" />
  </div>

  <v-tabs v-model="tab" class="mb-2" color="primary">
    <v-tab value="retail">Retail</v-tab>
    <v-tab value="wholesale">Wholesale</v-tab>
  </v-tabs>

  <ServerSideTable
    :key="tab"
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="clients"
  >
    <template #item.name="{ item }">
      <router-link :to="`/clients/${item.id}`">{{ item.name }}</router-link>
    </template>
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEdit(item)" />
      <v-icon-btn
        color="error"
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
  import { computed, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useClients } from '@/composables/useClients'
  import { API_BASE } from '@/config'

  const { createClient, updateClient, deleteClient } = useClients()

  const tab = ref('retail')
  const apiURL = computed(() => `${API_BASE}/clients/?client_type=${tab.value}`)
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Phone', key: 'phone', align: 'start' },
    { title: 'Loyalty Points', key: 'valid_loyalty_points', align: 'end' },
    { title: 'Created At', key: 'created_at', align: 'end' },
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
      phone (value) {
        if (/^[0-9-]{7,}$/.test(value ?? '')) return true
        return 'Phone number needs to be at least 7 digits.'
      },
    },
  })

  const name = useField('name')
  const phone = useField('phone')
  const clientType = ref('retail')

  function openCreate () {
    editingId.value = null
    handleReset()
    // Default to whichever tab is open, so adding from the Wholesale tab
    // doesn't silently create a retail client.
    clientType.value = tab.value
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    clientType.value = item.client_type ?? 'retail'
    setValues({ name: item.name, phone: item.phone })
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
        ? updateClient({ id: editingId.value, name: values.name, phone: values.phone, client_type: clientType.value })
        : createClient({ name: values.name, phone: values.phone, client_type: clientType.value }))
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
      await deleteClient(deleteTarget.value.id)
      deleteDialog.value = false
      tableRef.value?.reload()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }
</script>
