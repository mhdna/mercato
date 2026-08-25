<template>
  <v-dialog v-model="dialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>Add Loan</v-card-title>
      <v-card-text>
        <form @submit.prevent="submit">
          <v-text-field
            v-model="description.value.value"
            density="compact"
            :error-messages="description.errorMessage.value"
            label="Description"
          />
          <v-select
            v-model="categoryId.value.value"
            density="compact"
            :error-messages="categoryId.errorMessage.value"
            item-title="name"
            item-value="id"
            :items="activeCategories"
            label="Category"
          />
          <v-text-field
            v-model.number="amount.value.value"
            density="compact"
            :error-messages="amount.errorMessage.value"
            label="Amount"
            type="number"
          />
          <v-text-field
            v-model="currencyCode.value.value"
            density="compact"
            :error-messages="currencyCode.errorMessage.value"
            label="Currency Code"
          />

          <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
            {{ submitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeDialog" />
            <v-btn color="primary" :loading="submitting" text="Add Loan" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="categoryDialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>{{ editingCategoryId ? 'Edit Category' : 'Add a New Category' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitCategory">
          <v-text-field
            v-model="categoryName.value.value"
            density="compact"
            :error-messages="categoryName.errorMessage.value"
            hint="Lender source, e.g. Owner, External"
            label="Name"
            persistent-hint
          />
          <v-switch
            v-model="categoryIsActive.value.value"
            color="primary"
            density="compact"
            hide-details
            label="Active"
          />

          <v-alert v-if="categorySubmitError" class="mb-4 mt-4" type="error" variant="tonal">
            {{ categorySubmitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeCategoryDialog" />
            <v-btn color="primary" :loading="categorySubmitting" :text="editingCategoryId ? 'Save' : 'Add Category'" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-container>
    <v-card>
      <div class="d-flex justify-space-between align-center pe-4">
        <v-tabs v-model="tab" color="primary">
          <v-tab value="own">Loans</v-tab>
          <v-tab value="branch">Branch Loans</v-tab>
          <v-tab value="categories">Categories</v-tab>
        </v-tabs>
        <v-btn
          v-if="tab === 'own'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Loan"
          @click="openCreate"
        />
        <v-btn
          v-else-if="tab === 'categories'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Category"
          @click="openCreateCategory"
        />
        <v-select
          v-else
          v-model="branchFilter"
          clearable
          density="compact"
          hide-details
          item-title="name"
          item-value="id"
          :items="branches"
          placeholder="All branches"
          style="max-width: 260px"
          variant="outlined"
          @update:model-value="onBranchFilterChange"
        />
      </div>

      <v-window v-model="tab">
        <v-window-item value="own">
          <v-card-text>
            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="`${API_BASE}/loans?origin=central_loan`"
              :headers="headers"
              root-key="loans"
            >
              <template #item.category_id="{ item }">
                {{ categoryNameFor(item.category_id) }}
              </template>
              <template #item.amount="{ item }">
                {{ formatMoney(item.amount) }} {{ item.currency_code }}
              </template>
              <template #item.occurred_at="{ item }">
                {{ new Date(item.occurred_at).toLocaleString() }}
              </template>
            </ServerSideTable>
          </v-card-text>
        </v-window-item>

        <v-window-item value="branch">
          <v-card-text>
            <ServerSideTable
              ref="branchTableRef"
              :api-u-r-l="buildBranchLoansUrl()"
              :headers="branchHeaders"
              root-key="loans"
            >
              <template #item.branch_id="{ item }">
                {{ branchName(item.branch_id) }}
              </template>
              <template #item.category_id="{ item }">
                {{ categoryNameFor(item.category_id) }}
              </template>
              <template #item.amount="{ item }">
                {{ formatMoney(item.amount) }} {{ item.currency_code }}
              </template>
              <template #item.occurred_at="{ item }">
                {{ new Date(item.occurred_at).toLocaleString() }}
              </template>
              <template #item.received_at="{ item }">
                {{ new Date(item.received_at).toLocaleString() }}
              </template>
            </ServerSideTable>
          </v-card-text>
        </v-window-item>

        <v-window-item value="categories">
          <v-card-text>
            <v-table density="compact">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Status</th>
                  <th class="text-end">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="category in loanCategories" :key="category.id">
                  <td>{{ category.name }}</td>
                  <td>
                    <v-chip :color="category.is_active ? 'success' : 'default'" size="small">
                      {{ category.is_active ? 'Active' : 'Inactive' }}
                    </v-chip>
                  </td>
                  <td class="text-end">
                    <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEditCategory(category)" />
                    <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="removeCategory(category)" />
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-window-item>
      </v-window>
    </v-card>
  </v-container>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { computed, onMounted, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { authFetch } from '@/composables/useApi'
  import { useBranches } from '@/composables/useBranches'
  import { useLoanCategories } from '@/composables/useLoanCategories'
  import { API_BASE } from '@/config'

  const { branches, fetchBranches } = useBranches()
  const { loanCategories, fetchLoanCategories, createLoanCategory, updateLoanCategory, deleteLoanCategory } = useLoanCategories()
  onMounted(() => {
    fetchBranches()
    fetchLoanCategories()
  })

  const activeCategories = computed(() => loanCategories.value.filter(c => c.is_active))

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  function categoryNameFor (id) {
    return loanCategories.value.find(c => c.id === id)?.name ?? `Category #${id}`
  }

  function formatMoney (cents) {
    return (Number(cents) / 100).toFixed(2)
  }

  const tab = ref('own')

  const headers = [
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Category', key: 'category_id', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Date', key: 'occurred_at', align: 'start' },
  ]

  const branchHeaders = [
    { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
    { title: 'Description', key: 'description', align: 'start', sortable: false },
    { title: 'Category', key: 'category_id', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end', sortable: false },
    { title: 'Occurred At', key: 'occurred_at', align: 'start', sortable: false },
    { title: 'Received At', key: 'received_at', align: 'start', sortable: false },
  ]

  const tableRef = ref(null)
  const branchTableRef = ref(null)
  const branchFilter = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  function buildBranchLoansUrl () {
    const url = new URL(`${API_BASE}/loans`, window.location.origin)
    url.searchParams.set('origin', 'branch_loan')
    if (branchFilter.value) url.searchParams.set('branch_id', branchFilter.value)
    return url.toString()
  }

  function onBranchFilterChange () {
    branchTableRef.value?.reload()
  }

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      description (value) {
        if (value?.length >= 2) return true
        return 'Description is required.'
      },
      categoryId (value) {
        if (value) return true
        return 'Category is required.'
      },
      amount (value) {
        if (value !== undefined && value !== null && value !== '' && Number(value) > 0) return true
        return 'Amount must be greater than 0.'
      },
      currencyCode (value) {
        if (value?.length >= 2) return true
        return 'Currency code is required.'
      },
    },
  })

  const description = useField('description')
  const categoryId = useField('categoryId')
  const amount = useField('amount')
  const currencyCode = useField('currencyCode')

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
      const res = await authFetch(`${API_BASE}/loans`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          description: values.description,
          category_id: values.categoryId,
          amount: Math.round(Number(values.amount) * 100),
          currency_code: values.currencyCode,
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

  // -- Categories --

  const categoryDialog = ref(false)
  const categorySubmitting = ref(false)
  const categorySubmitError = ref('')
  const editingCategoryId = ref(null)

  const { handleSubmit: handleCategorySubmit, handleReset: handleCategoryReset, setValues: setCategoryValues } = useForm({
    validationSchema: {
      categoryName (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
    },
  })

  const categoryName = useField('categoryName')
  const categoryIsActive = useField('categoryIsActive')

  function openCreateCategory () {
    editingCategoryId.value = null
    handleCategoryReset()
    categoryIsActive.value.value = true
    categoryDialog.value = true
  }

  function openEditCategory (category) {
    editingCategoryId.value = category.id
    setCategoryValues({
      categoryName: category.name,
      categoryIsActive: category.is_active,
    })
    categoryDialog.value = true
  }

  function closeCategoryDialog () {
    categoryDialog.value = false
    handleCategoryReset()
    categorySubmitError.value = ''
    editingCategoryId.value = null
  }

  const submitCategory = handleCategorySubmit(async values => {
    categorySubmitting.value = true
    categorySubmitError.value = ''
    try {
      await (editingCategoryId.value
        ? updateLoanCategory({
          id: editingCategoryId.value,
          name: values.categoryName,
          is_active: !!values.categoryIsActive,
        })
        : createLoanCategory({
          name: values.categoryName,
          is_active: !!values.categoryIsActive,
        }))
      closeCategoryDialog()
    } catch (error) {
      categorySubmitError.value = error.message
    } finally {
      categorySubmitting.value = false
    }
  })

  async function removeCategory (category) {
    if (!confirm(`Delete category "${category.name}"? Categories still used by a loan can't be deleted.`)) return
    try {
      await deleteLoanCategory(category.id)
    } catch (error) {
      categorySubmitError.value = error.message
    }
  }
</script>
