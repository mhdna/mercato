<template>
  <v-dialog v-model="dialog" max-width="480">
    <v-card class="px-4">
      <v-card-title>Add Expense</v-card-title>
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
            clearable
            density="compact"
            :error-messages="categoryId.errorMessage.value"
            item-title="name"
            item-value="id"
            :items="activeExpenseCategories"
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
            <v-btn color="primary" :loading="submitting" text="Add Expense" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="recurringDialog" max-width="480">
    <v-card class="px-4">
      <v-card-title>Add Recurring Expense</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitRecurring">
          <v-text-field
            v-model="rDescription.value.value"
            density="compact"
            :error-messages="rDescription.errorMessage.value"
            label="Description"
          />
          <v-select
            v-model="rCategoryId.value.value"
            clearable
            density="compact"
            :error-messages="rCategoryId.errorMessage.value"
            item-title="name"
            item-value="id"
            :items="activeExpenseCategories"
            label="Category"
          />
          <v-text-field
            v-model.number="rAmount.value.value"
            density="compact"
            :error-messages="rAmount.errorMessage.value"
            label="Amount"
            type="number"
          />
          <v-text-field
            v-model="rCurrencyCode.value.value"
            density="compact"
            :error-messages="rCurrencyCode.errorMessage.value"
            label="Currency Code"
          />
          <v-row dense>
            <v-col cols="6">
              <v-text-field
                v-model.number="rIntervalCount.value.value"
                density="compact"
                :error-messages="rIntervalCount.errorMessage.value"
                label="Every"
                type="number"
              />
            </v-col>
            <v-col cols="6">
              <v-select
                v-model="rIntervalUnit.value.value"
                density="compact"
                :error-messages="rIntervalUnit.errorMessage.value"
                :items="[{ title: 'Days', value: 'day' }, { title: 'Months', value: 'month' }]"
                label="Unit"
              />
            </v-col>
          </v-row>
          <v-text-field
            v-model="rNextDueAt.value.value"
            density="compact"
            :error-messages="rNextDueAt.errorMessage.value"
            label="First Due Date"
            type="date"
          />

          <v-alert v-if="recurringSubmitError" class="mb-4" type="error" variant="tonal">
            {{ recurringSubmitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="closeRecurringDialog" />
            <v-btn color="primary" :loading="recurringSubmitting" text="Add Recurring Expense" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="categoryDialog" max-width="480">
    <v-card class="px-4">
      <v-card-title>{{ editingCategoryId ? 'Edit Category' : 'Add a New Category' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitCategory">
          <v-text-field
            v-model="categoryName.value.value"
            density="compact"
            :error-messages="categoryName.errorMessage.value"
            label="Name"
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
          <v-tab value="own">Expenses</v-tab>
          <v-tab value="recurring">Recurring</v-tab>
          <v-tab value="branch">Branch Expenses</v-tab>
          <v-tab value="categories">Categories</v-tab>
        </v-tabs>
        <v-btn
          v-if="tab === 'own'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Expense"
          @click="dialog = true"
        />
        <v-btn
          v-else-if="tab === 'recurring'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Recurring Expense"
          @click="recurringDialog = true"
        />
        <v-btn
          v-else-if="tab === 'categories'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Category"
          @click="openCreateCategory"
        />
        <v-select
          v-else-if="tab === 'branch'"
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
              :api-u-r-l="`${API_BASE}/expenses`"
              :headers="headers"
              root-key="expenses"
            >
              <template #item.category_id="{ item }">
                {{ categoryNameFor(item.category_id) }}
              </template>
            </ServerSideTable>
          </v-card-text>
        </v-window-item>

        <v-window-item value="recurring">
          <v-card-text>
            <v-alert v-if="recurringLoadError" class="mb-2" type="error" variant="tonal">
              {{ recurringLoadError }}
            </v-alert>
            <v-data-table density="compact" :headers="recurringHeaders" :items="recurringExpenses" :loading="recurringLoading">
              <template #item.category_id="{ item }">
                {{ categoryNameFor(item.category_id) }}
              </template>
              <template #item.amount="{ item }">
                {{ formatMoney(item.amount) }} {{ item.currency_code }}
              </template>
              <template #item.interval="{ item }">
                Every {{ item.interval_count }} {{ item.interval_unit }}{{ item.interval_count > 1 ? 's' : '' }}
              </template>
              <template #item.next_due_at="{ item }">
                {{ new Date(item.next_due_at).toLocaleDateString() }}
              </template>
              <template #item.active="{ item }">
                <v-switch
                  color="primary"
                  density="compact"
                  hide-details
                  :model-value="item.active"
                  @update:model-value="value => toggleRecurringActive(item, value)"
                />
              </template>
            </v-data-table>
          </v-card-text>
        </v-window-item>

        <v-window-item value="branch">
          <v-card-text>
            <ServerSideTable
              ref="branchTableRef"
              :api-u-r-l="buildBranchExpensesUrl()"
              :headers="branchHeaders"
              root-key="branch_expenses"
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
                <tr v-for="category in expenseCategories" :key="category.id">
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
  import { computed, onMounted, ref, watch } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { authFetch } from '@/composables/useApi'
  import { useBranches } from '@/composables/useBranches'
  import { useExpenseCategories } from '@/composables/useExpenseCategories'
  import { API_BASE } from '@/config'

  const { branches, fetchBranches } = useBranches()
  const { expenseCategories, fetchExpenseCategories, createExpenseCategory, updateExpenseCategory, deleteExpenseCategory } = useExpenseCategories()
  onMounted(() => {
    fetchBranches()
    fetchExpenseCategories()
  })

  const activeExpenseCategories = computed(() => expenseCategories.value.filter(c => c.is_active))

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  function categoryNameFor (id) {
    if (!id) return ''
    return expenseCategories.value.find(c => c.id === id)?.name ?? `Category #${id}`
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
    { title: 'Date', key: 'created_at', align: 'start' },
  ]

  const branchHeaders = [
    { title: 'Branch', key: 'branch_id', align: 'start', sortable: false },
    { title: 'Description', key: 'description', align: 'start', sortable: false },
    { title: 'Category', key: 'category_id', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end', sortable: false },
    { title: 'Occurred At', key: 'occurred_at', align: 'start', sortable: false },
    { title: 'Received At', key: 'received_at', align: 'start', sortable: false },
  ]

  const recurringHeaders = [
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Category', key: 'category_id', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Interval', key: 'interval', align: 'start', sortable: false },
    { title: 'Next Due', key: 'next_due_at', align: 'start' },
    { title: 'Active', key: 'active', align: 'center', sortable: false },
  ]

  const tableRef = ref(null)
  const branchTableRef = ref(null)
  const branchFilter = ref(null)
  const dialog = ref(false)
  const submitting = ref(false)
  const submitError = ref('')

  function buildBranchExpensesUrl () {
    const url = new URL(`${API_BASE}/branch_expenses`, window.location.origin)
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

  function closeDialog () {
    dialog.value = false
    handleReset()
    submitError.value = ''
  }

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/expenses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          description: values.description,
          category_id: values.categoryId ?? undefined,
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

  // -- Recurring expenses --

  const recurringExpenses = ref([])
  const recurringLoading = ref(false)
  const recurringLoadError = ref('')

  async function fetchRecurringExpenses () {
    recurringLoading.value = true
    recurringLoadError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/recurring_expenses`)
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      recurringExpenses.value = await res.json()
    } catch (error) {
      recurringLoadError.value = error.message
    } finally {
      recurringLoading.value = false
    }
  }

  watch(tab, value => {
    if (value === 'recurring' && recurringExpenses.value.length === 0) fetchRecurringExpenses()
  })

  async function toggleRecurringActive (item, active) {
    try {
      const res = await authFetch(`${API_BASE}/recurring_expenses/active`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: item.id, active }),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }
      item.active = active
    } catch (error) {
      recurringLoadError.value = error.message
    }
  }

  const recurringDialog = ref(false)
  const recurringSubmitting = ref(false)
  const recurringSubmitError = ref('')

  const { handleSubmit: handleRecurringSubmit, handleReset: handleRecurringReset } = useForm({
    validationSchema: {
      rDescription (value) {
        if (value?.length >= 2) return true
        return 'Description is required.'
      },
      rAmount (value) {
        if (value !== undefined && value !== null && value !== '' && Number(value) > 0) return true
        return 'Amount must be greater than 0.'
      },
      rCurrencyCode (value) {
        if (value?.length >= 2) return true
        return 'Currency code is required.'
      },
      rIntervalCount (value) {
        if (value !== undefined && value !== null && value !== '' && Number(value) > 0) return true
        return 'Must be greater than 0.'
      },
      rIntervalUnit (value) {
        if (value === 'day' || value === 'month') return true
        return 'Pick an interval unit.'
      },
      rNextDueAt (value) {
        if (value) return true
        return 'First due date is required.'
      },
    },
  })

  const rDescription = useField('rDescription')
  const rCategoryId = useField('rCategoryId')
  const rAmount = useField('rAmount')
  const rCurrencyCode = useField('rCurrencyCode')
  const rIntervalCount = useField('rIntervalCount', undefined, { initialValue: 1 })
  const rIntervalUnit = useField('rIntervalUnit', undefined, { initialValue: 'month' })
  const rNextDueAt = useField('rNextDueAt')

  function closeRecurringDialog () {
    recurringDialog.value = false
    handleRecurringReset()
    recurringSubmitError.value = ''
  }

  const submitRecurring = handleRecurringSubmit(async values => {
    recurringSubmitting.value = true
    recurringSubmitError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/recurring_expenses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          description: values.rDescription,
          category_id: values.rCategoryId ?? undefined,
          amount: Math.round(Number(values.rAmount) * 100),
          currency_code: values.rCurrencyCode,
          interval_count: Number(values.rIntervalCount),
          interval_unit: values.rIntervalUnit,
          next_due_at: new Date(values.rNextDueAt).toISOString(),
        }),
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }

      closeRecurringDialog()
      fetchRecurringExpenses()
    } catch (error) {
      recurringSubmitError.value = error.message
    } finally {
      recurringSubmitting.value = false
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
        ? updateExpenseCategory({
          id: editingCategoryId.value,
          name: values.categoryName,
          is_active: !!values.categoryIsActive,
        })
        : createExpenseCategory({
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
    if (!confirm(`Delete category "${category.name}"? Categories still used by an expense can't be deleted.`)) return
    try {
      await deleteExpenseCategory(category.id)
    } catch (error) {
      categorySubmitError.value = error.message
    }
  }
</script>
