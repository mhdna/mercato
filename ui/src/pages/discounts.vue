<template>
  <v-dialog v-model="dialog" max-width="560" :persistent="true">
    <v-card class="px-4">
      <v-card-title>{{ editingId ? 'Edit Discount List' : 'Add a New Discount List' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="name.value.value"
            density="compact"
            :error-messages="name.errorMessage.value"
            label="Name"
          />
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model="validFrom.value.value"
                density="compact"
                :error-messages="validFrom.errorMessage.value"
                label="Valid From"
                type="date"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model="validTo.value.value"
                density="compact"
                :error-messages="validTo.errorMessage.value"
                label="Valid To"
                type="date"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="6">
              <v-switch v-model="isActive.value.value" density="compact" hide-details label="Active" />
            </v-col>
            <v-col cols="6">
              <v-switch v-model="isDefault.value.value" density="compact" hide-details label="Default" />
            </v-col>
          </v-row>

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" :text="editingId ? 'Save' : 'Add List'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="itemsDialog" max-width="640">
    <v-card class="px-4">
      <v-card-title>Products in "{{ itemsTarget?.name }}"</v-card-title>
      <v-card-text>
        <div class="d-flex align-center mb-4" style="gap: 12px;">
          <v-select
            v-model="itemProductId"
            density="compact"
            hide-details
            :item-title="p => `${p.code} - ${p.name}`"
            :item-value="p => p.id"
            :items="products"
            label="Product"
          />
          <v-text-field
            v-model.number="itemDiscount"
            density="compact"
            hide-details
            label="Discount %"
            style="max-width: 140px;"
            type="number"
          />
          <v-btn :loading="itemSubmitting" text="Add" @click="addItem" />
        </div>
        <v-alert v-if="itemError" class="mb-4" type="error" variant="tonal">{{ itemError }}</v-alert>
        <v-table density="compact">
          <thead>
            <tr>
              <th>Product ID</th>
              <th>Discount %</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.product_id">
              <td>{{ item.product_id }}</td>
              <td>{{ item.discount }}</td>
              <td class="text-end">
                <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="removeItem(item)" />
              </td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="itemsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Discount Lists</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Discount List" @click="openCreate" />
  </div>

  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :headers="headers"
    :max-page-size="10"
    root-key="discount_lists"
  >
    <template #item.is_active="{ item }">
      <v-icon :color="item.is_active ? 'success' : 'grey'" :icon="item.is_active ? 'mdi-check-circle' : 'mdi-close-circle'" />
    </template>
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-format-list-bulleted" size="small" variant="text" @click="openItems(item)" />
      <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEdit(item)" />
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useDiscountLists } from '@/composables/useDiscountLists'
  import { useProducts } from '@/composables/useProducts'
  import { API_BASE } from '@/config'

  const { createDiscountList, updateDiscountList, fetchDiscountListItems, createDiscountListItem, deleteDiscountListItem } = useDiscountLists()
  const { products, fetchProducts } = useProducts()
  fetchProducts()

  const apiURL = `${API_BASE}/discount_lists`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Active', key: 'is_active', align: 'center' },
    { title: 'Valid From', key: 'valid_from', align: 'end' },
    { title: 'Valid To', key: 'valid_to', align: 'end' },
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
      validFrom (value) {
        return !!value || 'Valid from date is required.'
      },
      validTo (value) {
        return !!value || 'Valid to date is required.'
      },
      isActive () {
        return true
      },
      isDefault () {
        return true
      },
    },
  })

  const name = useField('name')
  const validFrom = useField('validFrom')
  const validTo = useField('validTo')
  const isActive = useField('isActive')
  const isDefault = useField('isDefault')

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
      validFrom: item.valid_from?.slice(0, 10),
      validTo: item.valid_to?.slice(0, 10),
      isActive: item.is_active,
      isDefault: item.is_default,
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
      const payload = {
        name: values.name,
        is_active: !!values.isActive,
        is_default: !!values.isDefault,
        valid_from: new Date(values.validFrom).toISOString(),
        valid_to: new Date(values.validTo).toISOString(),
      }
      await (editingId.value ? updateDiscountList({ id: editingId.value, ...payload }) : createDiscountList(payload))
      closeDialog()
      tableRef.value?.reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })

  const itemsDialog = ref(false)
  const itemsTarget = ref(null)
  const items = ref([])
  const itemProductId = ref(null)
  const itemDiscount = ref(null)
  const itemSubmitting = ref(false)
  const itemError = ref('')

  async function openItems (item) {
    itemsTarget.value = item
    itemError.value = ''
    itemsDialog.value = true
    try {
      items.value = await fetchDiscountListItems(item.id)
    } catch (error) {
      itemError.value = error.message
    }
  }

  async function addItem () {
    if (!itemProductId.value || !itemDiscount.value) return
    itemSubmitting.value = true
    itemError.value = ''
    try {
      await createDiscountListItem({
        discountListId: itemsTarget.value.id,
        productId: itemProductId.value,
        discount: itemDiscount.value,
      })
      items.value = await fetchDiscountListItems(itemsTarget.value.id)
      itemProductId.value = null
      itemDiscount.value = null
    } catch (error) {
      itemError.value = error.message
    } finally {
      itemSubmitting.value = false
    }
  }

  async function removeItem (item) {
    try {
      await deleteDiscountListItem(itemsTarget.value.id, item.product_id)
      items.value = await fetchDiscountListItems(itemsTarget.value.id)
    } catch (error) {
      itemError.value = error.message
    }
  }
</script>
