<template>
  <v-dialog v-model="dialog" max-width="560">
    <v-card class="px-4">
      <v-card-title>Add a New Supplier</v-card-title>
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
          <v-text-field
            v-model="country.value.value"
            density="compact"
            :error-messages="country.errorMessage.value"
            label="Country"
          />
          <v-text-field
            v-model="address.value.value"
            density="compact"
            :error-messages="address.errorMessage.value"
            label="Address"
          />
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model.number="latitude.value.value"
                density="compact"
                :error-messages="latitude.errorMessage.value"
                label="Latitude"
                type="number"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="longitude.value.value"
                density="compact"
                :error-messages="longitude.errorMessage.value"
                label="Longitude"
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
            <v-btn color="primary" :loading="submitting" text="Add Supplier" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="suppliers-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-truck-delivery" />
        <span>Suppliers</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="supplier-search"
          clearable
          density="compact"
          hide-details
          label="Search suppliers"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Supplier"
          variant="flat"
          @click="dialog = true"
        />
      </v-card-title>
      <v-divider />

      <ServerSideTable
        ref="tableRef"
        :api-u-r-l="apiURL"
        density="comfortable"
        :external-search="search"
        flush
        :headers="headers"
        root-key="suppliers"
        :show-search-icon="false"
      />
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { authFetch } from '@/composables/useApi'
  import { API_BASE } from '@/config'

  const apiURL = `${API_BASE}/suppliers`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Phone', key: 'phone', align: 'start' },
    { title: 'Country', key: 'country', align: 'start' },
    { title: 'Address', key: 'address', align: 'start' },
    { title: 'Created At', key: 'created_at', align: 'end' },
  ])

  const tableRef = ref(null)
  const search = ref('')
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      name (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
      phone (value) {
        if (/^[0-9-]{7,}$/.test(value)) return true
        return 'Phone number needs to be at least 7 digits.'
      },
      country (value) {
        if (value?.length >= 2) return true
        return 'Country is required.'
      },
      address (value) {
        if (value?.length >= 2) return true
        return 'Address is required.'
      },
      latitude (value) {
        if (value !== undefined && value !== null && value !== '') return true
        return 'Latitude is required.'
      },
      longitude (value) {
        if (value !== undefined && value !== null && value !== '') return true
        return 'Longitude is required.'
      },
    },
  })

  const name = useField('name')
  const phone = useField('phone')
  const country = useField('country')
  const address = useField('address')
  const latitude = useField('latitude')
  const longitude = useField('longitude')

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const res = await authFetch(apiURL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: values.name,
          phone: values.phone,
          country: values.country,
          address: values.address,
          latitude: values.latitude,
          longitude: values.longitude,
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
</script>

<style scoped>
.page-root {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.suppliers-card {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.supplier-search {
  flex: 0 1 320px;
}
</style>
