<template>
  <v-dialog v-model="dialog" max-width="560">
    <v-card class="px-4">
      <v-card-title>{{ editingId ? 'Edit Inventory' : 'Add a New Inventory' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
          />
          <v-select
            v-model="type.value.value"
            density="compact"
            :disabled="!!editingId"
            :error-messages="type.errorMessage.value"
            :items="['warehouse', 'store']"
            label="Type"
          />
          <v-text-field
            v-model="code.value.value"
            density="compact"
            :disabled="!!editingId"
            :error-messages="code.errorMessage.value"
            label="Code"
          />
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model.number="latitude.value.value"
                clearable
                density="compact"
                :disabled="!!editingId"
                label="Latitude (optional)"
                type="number"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="longitude.value.value"
                clearable
                density="compact"
                :disabled="!!editingId"
                label="Longitude (optional)"
                type="number"
              />
            </v-col>
          </v-row>

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingId ? 'Save' : 'Add Inventory'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="deleteDialog" max-width="420">
    <v-card>
      <v-card-title>Delete Inventory</v-card-title>
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
    <h2 class="text-h6">Inventory</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Inventory" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="inventories"
  >
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
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { API_BASE } from '@/config'

  const { createInventory, updateInventory, deleteInventory } = useInventories()

  const apiURL = `${API_BASE}/inventories/`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Type', key: 'type', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
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
      type (value) {
        if (editingId.value) return true
        return ['warehouse', 'store'].includes(value) || 'Select a type.'
      },
      code (value) {
        if (editingId.value) return true
        return (value?.length >= 1) || 'Code is required.'
      },
    },
  })

  const name = useField('name')
  const type = useField('type')
  const code = useField('code')
  const latitude = useField('latitude')
  const longitude = useField('longitude')

  function openCreate () {
    editingId.value = null
    handleReset()
    dialog.value = true
  }

  function openEdit (item) {
    editingId.value = item.id
    setValues({ name: item.name, type: item.type, code: item.code })
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
        ? updateInventory({ id: editingId.value, name: values.name })
        : createInventory({
          name: values.name,
          type: values.type,
          code: values.code,
          longitude: values.longitude || null,
          latitude: values.latitude || null,
        }))
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
      await deleteInventory(deleteTarget.value.id)
      deleteDialog.value = false
      tableRef.value?.reload()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }
</script>
