<template>
  <!-- Add / edit a one-off expense -->
  <v-dialog v-model="expenseDialog" max-width="460">
    <v-card class="px-4">
      <v-card-title>{{ editingExpenseId ? 'Edit Expense' : 'Add Expense' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitExpense">
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
          >
            <template #item="{ props: itemProps, item }">
              <v-list-item v-bind="itemProps" :prepend-icon="categoryIcon(item.raw)" />
            </template>
          </v-select>
          <v-row density="compact">
            <v-col cols="7">
              <v-text-field
                v-model.number="amount.value.value"
                density="compact"
                :error-messages="amount.errorMessage.value"
                label="Amount"
                type="number"
              />
            </v-col>
            <v-col cols="5">
              <CurrencySelect
                v-model="currencyCode.value.value"
                :error-messages="currencyCode.errorMessage.value"
              />
            </v-col>
          </v-row>

          <v-alert v-if="expenseSubmitError" class="mb-4" type="error" variant="tonal">
            {{ expenseSubmitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-btn
              v-if="editingExpenseId"
              color="error"
              prepend-icon="mdi-delete"
              text="Delete"
              variant="text"
              @click="deleteCurrentExpense"
            />
            <v-spacer />
            <v-btn text="Cancel" @click="closeExpenseDialog" />
            <v-btn
              color="primary"
              :loading="expenseSubmitting"
              :text="editingExpenseId ? 'Save' : 'Add Expense'"
              type="submit"
            />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <!-- Add a recurring expense template -->
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
          <v-row density="compact">
            <v-col cols="7">
              <v-text-field
                v-model.number="rAmount.value.value"
                density="compact"
                :error-messages="rAmount.errorMessage.value"
                label="Amount"
                type="number"
              />
            </v-col>
            <v-col cols="5">
              <CurrencySelect
                v-model="rCurrencyCode.value.value"
                :error-messages="rCurrencyCode.errorMessage.value"
              />
            </v-col>
          </v-row>
          <v-row density="compact">
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

  <!-- Record an expense against a branch (central-side entry) -->
  <v-dialog v-model="branchExpenseDialog" max-width="460">
    <v-card class="px-4">
      <v-card-title>Add Branch Expense</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitBranchExpense">
          <v-select
            v-model="bxForm.branchId"
            density="compact"
            :error-messages="bxErrors.branchId"
            item-title="name"
            item-value="id"
            :items="branches"
            label="Branch"
          />
          <v-text-field
            v-model="bxForm.description"
            density="compact"
            :error-messages="bxErrors.description"
            label="Description"
          />
          <v-select
            v-model="bxForm.categoryId"
            clearable
            density="compact"
            item-title="name"
            item-value="id"
            :items="activeExpenseCategories"
            label="Category"
          >
            <template #item="{ props: itemProps, item }">
              <v-list-item v-bind="itemProps" :prepend-icon="categoryIcon(item.raw)" />
            </template>
          </v-select>
          <v-row density="compact">
            <v-col cols="7">
              <v-text-field
                v-model.number="bxForm.amount"
                density="compact"
                :error-messages="bxErrors.amount"
                label="Amount"
                type="number"
              />
            </v-col>
            <v-col cols="5">
              <CurrencySelect v-model="bxForm.currencyCode" :error-messages="bxErrors.currencyCode" />
            </v-col>
          </v-row>
          <v-text-field
            v-model="bxForm.occurredAt"
            density="compact"
            hint="Defaults to now"
            label="Date"
            persistent-hint
            type="date"
          />

          <v-alert v-if="branchExpenseSubmitError" class="mb-4 mt-4" type="error" variant="tonal">
            {{ branchExpenseSubmitError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="branchExpenseDialog = false" />
            <v-btn color="primary" :loading="branchExpenseSubmitting" text="Add Branch Expense" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="expenses-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-cash-minus" />
        <span>Expenses</span>
        <v-spacer />
        <v-tabs v-model="tab" class="expense-tabs" color="primary" density="compact">
          <v-tab value="own">Mine</v-tab>
          <v-tab value="recurring">Recurring</v-tab>
          <v-tab value="branch">Branches</v-tab>
        </v-tabs>
        <v-text-field
          v-model="search"
          class="expense-search"
          clearable
          density="compact"
          hide-details
          label="Search expenses"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <div class="d-flex ga-2">
          <v-btn
            v-if="tab === 'recurring'"
            color="primary"
            prepend-icon="mdi-plus"
            text="Add Recurring"
            @click="recurringDialog = true"
          />
          <v-btn
            v-else-if="tab === 'own'"
            color="primary"
            prepend-icon="mdi-plus"
            text="Add Expense"
            @click="openCreateExpense"
          />
          <v-btn
            v-else-if="tab === 'branch'"
            color="primary"
            prepend-icon="mdi-plus"
            text="Add Branch Expense"
            @click="openCreateBranchExpense"
          />
        </div>
      </v-card-title>
      <v-divider />

      <div class="expenses-layout">
        <CategorySidebar
          v-model="selectedCategory"
          :categories="expenseCategories"
          @add="openCreateCategory"
          @delete="removeCategory"
          @edit="openEditCategory"
        />

        <v-divider vertical />

        <div class="expenses-content">
          <v-window v-model="tab">
            <!-- Mine -->
            <v-window-item value="own">
              <v-alert v-if="expensesError" class="ma-4" type="error" variant="tonal">{{ expensesError }}</v-alert>

              <div class="px-4 pt-3">
                <div v-if="filteredExpenses.length > 0" class="d-flex align-center ga-2 mb-1">
                  <v-checkbox-btn
                    :indeterminate="expenseSelected.length > 0 && !allExpensesSelected"
                    label="Select all"
                    :model-value="allExpensesSelected"
                    @update:model-value="toggleAllExpenses"
                  />
                </div>
                <BulkDeleteBar
                  :count="expenseSelected.length"
                  :error="expenseSelError"
                  :loading="expenseDeleting"
                  @clear="resetExpenseSel"
                  @confirm="bulkDeleteExpenses"
                />
              </div>

              <v-list v-if="filteredExpenses.length > 0" class="py-0" lines="two">
                <template v-for="(item, i) in filteredExpenses" :key="item.id">
                  <v-list-item class="expense-row px-4 py-2" @click="openEditExpense(item)">
                    <template #prepend>
                      <v-checkbox-btn
                        class="me-1"
                        :model-value="expenseSelected.includes(item.id)"
                        @click.stop
                        @update:model-value="toggleExpense(item.id)"
                      />
                      <v-avatar
                        class="me-3"
                        :color="expenseColor(item)"
                        rounded="lg"
                        size="44"
                        variant="tonal"
                      >
                        <v-icon :color="expenseColor(item)" :icon="expenseIcon(item)" size="22" />
                      </v-avatar>
                    </template>
                    <v-list-item-title class="font-weight-medium">{{ item.description }}</v-list-item-title>
                    <v-list-item-subtitle>{{ prettyDate(item.created_at) }} · {{ categoryLabel(item.category_id) }}</v-list-item-subtitle>
                    <template #append>
                      <span class="text-error font-weight-medium">−{{ money(item.amount) }} {{ item.currency_code }}</span>
                    </template>
                  </v-list-item>
                  <v-divider v-if="i < filteredExpenses.length - 1" />
                </template>
              </v-list>
              <div v-else-if="!expensesLoading" class="text-medium-emphasis text-center pa-10">No matching expenses.</div>

              <div class="d-flex justify-center pa-3">
                <v-progress-circular v-if="expensesLoading" color="primary" indeterminate size="24" />
                <div v-if="expensesHasMore" ref="expensesSentinel" class="infinite-scroll-sentinel" />
              </div>
            </v-window-item>

            <!-- Recurring -->
            <v-window-item value="recurring">
              <v-card-text>
                <v-alert v-if="recurringLoadError" class="mb-2" type="error" variant="tonal">
                  {{ recurringLoadError }}
                </v-alert>
                <v-data-table density="compact" :headers="recurringHeaders" :items="filteredRecurring" :loading="recurringLoading">
                  <template #item.category_id="{ item }">
                    <CategoryChip :category="categoryFor(item.category_id)" />
                  </template>
                  <template #item.amount="{ item }">
                    {{ money(item.amount) }} {{ item.currency_code }}
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

            <!-- Branches -->
            <v-window-item value="branch">
              <div class="d-flex px-4 pt-3">
                <v-select
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
                  @update:model-value="reloadBranchExpenses"
                />
              </div>

              <v-alert v-if="branchError" class="ma-4" type="error" variant="tonal">{{ branchError }}</v-alert>

              <v-list v-if="filteredBranchExpenses.length > 0" class="py-0" lines="two">
                <template v-for="(item, i) in filteredBranchExpenses" :key="item.id">
                  <v-list-item class="px-4 py-2">
                    <template #prepend>
                      <v-avatar
                        class="me-3"
                        :color="expenseColor(item)"
                        rounded="lg"
                        size="44"
                        variant="tonal"
                      >
                        <v-icon :color="expenseColor(item)" :icon="expenseIcon(item)" size="22" />
                      </v-avatar>
                    </template>
                    <v-list-item-title class="font-weight-medium">{{ item.description }}</v-list-item-title>
                    <v-list-item-subtitle>
                      {{ prettyDate(item.occurred_at) }} · {{ branchName(item.branch_id) }} · {{ categoryLabel(item.category_id) }}
                    </v-list-item-subtitle>
                    <template #append>
                      <span class="text-error font-weight-medium">−{{ money(item.amount) }} {{ item.currency_code }}</span>
                    </template>
                  </v-list-item>
                  <v-divider v-if="i < filteredBranchExpenses.length - 1" />
                </template>
              </v-list>
              <div v-else-if="!branchLoading" class="text-medium-emphasis text-center pa-10">No branch expenses yet.</div>

              <div class="d-flex justify-center pa-3">
                <v-progress-circular v-if="branchLoading" color="primary" indeterminate size="24" />
                <div v-if="branchHasMore" ref="branchSentinel" class="infinite-scroll-sentinel" />
              </div>
            </v-window-item>
          </v-window>
        </div>
      </div>
    </v-card>

    <CategoryDialog
      v-model="categoryDialog"
      :category="editingCategory"
      name-hint="e.g. Utilities, Rent, Payroll"
      :on-submit="submitCategory"
    />
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
  import CategoryChip from '@/components/CategoryChip.vue'
  import CategorySidebar from '@/components/CategorySidebar.vue'
  import CategoryDialog from '@/components/Forms/CategoryDialog.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import { authFetch } from '@/composables/useApi'
  import { useBranches } from '@/composables/useBranches'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useExpenseCategories } from '@/composables/useExpenseCategories'
  import { API_BASE } from '@/config'
  import { categoryIcon } from '@/data/expenseCategoryIcons'

  const PAGE_SIZE = 25

  const { branches, fetchBranches } = useBranches()
  const {
    expenseCategories,
    fetchExpenseCategories,
    createExpenseCategory,
    updateExpenseCategory,
    deleteExpenseCategory,
  } = useExpenseCategories()

  onMounted(() => {
    fetchBranches()
    fetchExpenseCategories()
    fetchExpenses()
  })

  const tab = ref('own')
  const search = ref('')
  const selectedCategory = ref(null)

  const activeExpenseCategories = computed(() => expenseCategories.value.filter(c => c.is_active))

  // Narrow any list to the category picked in the sidebar ("All" == null).
  function matchesCategory (item) {
    if (!selectedCategory.value) return true
    return catId(item.category_id) === selectedCategory.value
  }

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  // Go serialises a nullable category_id as sql.NullInt64 ({ Int64, Valid }),
  // but a plain number arrives from some paths too -- normalise both.
  function catId (raw) {
    if (raw && typeof raw === 'object') return raw.Valid ? raw.Int64 : null
    return raw || null
  }

  function categoryFor (raw) {
    const id = catId(raw)
    if (!id) return null
    return expenseCategories.value.find(c => c.id === id) ?? null
  }

  function categoryLabel (id) {
    return categoryFor(id)?.name ?? 'Uncategorised'
  }

  function expenseIcon (item) {
    return categoryFor(item.category_id)?.icon || 'mdi-cash-minus'
  }

  function expenseColor (item) {
    return categoryFor(item.category_id)?.color || 'red'
  }

  function money (cents) {
    return (Number(cents) / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }

  function prettyDate (value) {
    if (!value) return ''
    return new Date(value).toLocaleString(undefined, {
      day: 'numeric', month: 'short', hour: 'numeric', minute: '2-digit',
    })
  }

  async function apiJSON (url, options) {
    const res = await authFetch(url, options)
    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error(err.error || `Request failed with status ${res.status}`)
    }
    if (res.status === 204) return null
    return res.json().catch(() => null)
  }

  // -- Mine list --

  const expenses = ref([])
  const expensesLoading = ref(false)
  const expensesError = ref('')
  const expensesHasMore = ref(false)
  const expensesSentinel = ref(null)
  let expensesObserver = null

  const {
    selected: expenseSelected,
    deleting: expenseDeleting,
    error: expenseSelError,
    reset: resetExpenseSel,
    run: runExpenseBulk,
  } = useBulkDelete()

  const allExpensesSelected = computed(() =>
    filteredExpenses.value.length > 0
    && filteredExpenses.value.every(e => expenseSelected.value.includes(e.id)),
  )

  function toggleExpense (id) {
    const i = expenseSelected.value.indexOf(id)
    if (i === -1) expenseSelected.value.push(id)
    else expenseSelected.value.splice(i, 1)
  }

  function toggleAllExpenses () {
    expenseSelected.value = allExpensesSelected.value ? [] : filteredExpenses.value.map(e => e.id)
  }

  async function bulkDeleteExpenses () {
    try {
      await runExpenseBulk('/expenses/bulk_delete', { ids: expenseSelected.value })
      resetExpenseSel()
      fetchExpenses()
    } catch { /* error shown in the bar */ }
  }

  function deleteCurrentExpense () {
    const item = filteredExpenses.value.find(e => e.id === editingExpenseId.value)
    closeExpenseDialog()
    if (item) removeExpense(item)
  }

  // Search + category are applied server-side for own expenses (see
  // fetchExpenses), so this is a straight passthrough now.
  const filteredExpenses = computed(() => expenses.value)

  async function fetchExpenses (append = false) {
    expensesLoading.value = true
    expensesError.value = ''
    try {
      const offset = append ? expenses.value.length : 0
      const url = new URL(`${API_BASE}/expenses`, window.location.origin)
      url.searchParams.set('page_size', PAGE_SIZE)
      url.searchParams.set('page_id', offset)
      if (search.value.trim()) url.searchParams.set('search', search.value.trim())
      if (selectedCategory.value) url.searchParams.set('category_id', selectedCategory.value)
      const data = await apiJSON(url.toString())
      const rows = Array.isArray(data) ? data : []
      expenses.value = append ? [...expenses.value, ...rows] : rows
      expensesHasMore.value = rows.length === PAGE_SIZE
    } catch (error) {
      expensesError.value = error.message
    } finally {
      expensesLoading.value = false
      nextTick(setupInfiniteScroll)
    }
  }

  function loadMoreExpenses () {
    if (expensesLoading.value || !expensesHasMore.value) return
    fetchExpenses(true)
  }

  // -- Add / edit expense --

  const expenseDialog = ref(false)
  const editingExpenseId = ref(null)
  const expenseSubmitting = ref(false)
  const expenseSubmitError = ref('')

  const { handleSubmit, handleReset, setValues } = useForm({
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

  function openCreateExpense () {
    editingExpenseId.value = null
    handleReset()
    currencyCode.value.value = 'USD'
    expenseSubmitError.value = ''
    expenseDialog.value = true
  }

  function openEditExpense (item) {
    editingExpenseId.value = item.id
    setValues({
      description: item.description,
      categoryId: catId(item.category_id),
      amount: Number(item.amount) / 100,
      currencyCode: item.currency_code,
    })
    expenseSubmitError.value = ''
    expenseDialog.value = true
  }

  function closeExpenseDialog () {
    expenseDialog.value = false
    handleReset()
    editingExpenseId.value = null
    expenseSubmitError.value = ''
  }

  const submitExpense = handleSubmit(async values => {
    expenseSubmitting.value = true
    expenseSubmitError.value = ''
    try {
      const body = {
        description: values.description,
        category_id: values.categoryId || undefined,
        amount: Math.round(Number(values.amount) * 100),
        currency_code: values.currencyCode,
      }
      await (editingExpenseId.value
        ? apiJSON(`${API_BASE}/expenses`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id: editingExpenseId.value, ...body }),
        })
        : apiJSON(`${API_BASE}/expenses`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        }))
      closeExpenseDialog()
      fetchExpenses()
    } catch (error) {
      expenseSubmitError.value = error.message
    } finally {
      expenseSubmitting.value = false
    }
  })

  async function removeExpense (item) {
    if (!confirm(`Delete expense "${item.description}"?`)) return
    try {
      await apiJSON(`${API_BASE}/expenses/${item.id}`, { method: 'DELETE' })
      fetchExpenses()
    } catch (error) {
      expensesError.value = error.message
    }
  }

  // -- Branch expenses list --

  const branchExpenses = ref([])
  const branchLoading = ref(false)
  const branchError = ref('')
  const branchHasMore = ref(false)
  const branchFilter = ref(null)
  let branchLoaded = false
  const branchSentinel = ref(null)
  let branchObserver = null

  function setupInfiniteScroll () {
    expensesObserver?.disconnect()
    branchObserver?.disconnect()
    expensesObserver = new IntersectionObserver(entries => {
      if (entries[0]?.isIntersecting) loadMoreExpenses()
    }, { rootMargin: '160px' })
    branchObserver = new IntersectionObserver(entries => {
      if (entries[0]?.isIntersecting) loadMoreBranchExpenses()
    }, { rootMargin: '160px' })
    if (expensesSentinel.value) expensesObserver.observe(expensesSentinel.value)
    if (branchSentinel.value) branchObserver.observe(branchSentinel.value)
  }

  onMounted(() => setupInfiniteScroll())
  onUnmounted(() => {
    expensesObserver?.disconnect()
    branchObserver?.disconnect()
  })
  watch([expensesHasMore, branchHasMore], () => nextTick(setupInfiniteScroll))

  // Picking a different sidebar category invalidates the selection and,
  // for own expenses, refetches server-side from the first page.
  watch(selectedCategory, () => {
    resetExpenseSel()
    fetchExpenses(false)
  })

  // Debounced server-side search for both lists (recurring stays local —
  // it's a small, fully-loaded set).
  let searchTimer
  watch(search, () => {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      fetchExpenses(false)
      if (branchLoaded) fetchBranchExpenses(false)
    }, 300)
  })

  // Branch rows filter server-side on description (search) + branch; the
  // category narrowing stays client-side (no category_id on /branch_expenses).
  const filteredBranchExpenses = computed(() => branchExpenses.value.filter(item => matchesCategory(item)))

  async function fetchBranchExpenses (append = false) {
    branchLoading.value = true
    branchError.value = ''
    try {
      const offset = append ? branchExpenses.value.length : 0
      const url = new URL(`${API_BASE}/branch_expenses`, window.location.origin)
      url.searchParams.set('page_size', PAGE_SIZE)
      url.searchParams.set('page_id', offset)
      if (branchFilter.value) url.searchParams.set('branch_id', branchFilter.value)
      if (search.value.trim()) url.searchParams.set('search', search.value.trim())
      const data = await apiJSON(url.toString())
      const rows = data?.branch_expenses ?? []
      branchExpenses.value = append ? [...branchExpenses.value, ...rows] : rows
      branchHasMore.value = rows.length === PAGE_SIZE
      branchLoaded = true
    } catch (error) {
      branchError.value = error.message
    } finally {
      branchLoading.value = false
      nextTick(setupInfiniteScroll)
    }
  }

  function loadMoreBranchExpenses () {
    if (branchLoading.value || !branchHasMore.value) return
    fetchBranchExpenses(true)
  }

  function reloadBranchExpenses () {
    fetchBranchExpenses(false)
  }

  // -- Add a branch expense (central-side entry) --

  const branchExpenseDialog = ref(false)
  const branchExpenseSubmitting = ref(false)
  const branchExpenseSubmitError = ref('')
  const bxForm = ref({
    branchId: null,
    description: '',
    categoryId: null,
    amount: null,
    currencyCode: 'USD',
    occurredAt: '',
  })

  const bxErrors = computed(() => {
    const f = bxForm.value
    return {
      branchId: f.branchId ? '' : 'Pick a branch.',
      description: (f.description ?? '').trim().length >= 2 ? '' : 'Description is required.',
      amount: Number(f.amount) > 0 ? '' : 'Amount must be greater than 0.',
      currencyCode: (f.currencyCode ?? '').length >= 2 ? '' : 'Currency code is required.',
    }
  })

  function openCreateBranchExpense () {
    bxForm.value = {
      branchId: branchFilter.value ?? null,
      description: '',
      categoryId: null,
      amount: null,
      currencyCode: 'USD',
      occurredAt: '',
    }
    branchExpenseSubmitError.value = ''
    branchExpenseDialog.value = true
  }

  async function submitBranchExpense () {
    if (Object.values(bxErrors.value).some(Boolean)) return
    branchExpenseSubmitting.value = true
    branchExpenseSubmitError.value = ''
    try {
      const f = bxForm.value
      await apiJSON(`${API_BASE}/branch_expenses`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          branch_id: f.branchId,
          description: f.description.trim(),
          category_id: f.categoryId || undefined,
          amount: Math.round(Number(f.amount) * 100),
          currency_code: f.currencyCode,
          occurred_at: f.occurredAt ? new Date(f.occurredAt).toISOString() : undefined,
        }),
      })
      branchExpenseDialog.value = false
      branchLoaded = true
      fetchBranchExpenses(false)
    } catch (error) {
      branchExpenseSubmitError.value = error.message
    } finally {
      branchExpenseSubmitting.value = false
    }
  }

  watch(tab, value => {
    if (value === 'branch' && !branchLoaded) fetchBranchExpenses()
    if (value === 'recurring' && recurringExpenses.value.length === 0) fetchRecurringExpenses()
    nextTick(setupInfiniteScroll)
  })

  // -- Recurring expenses --

  const recurringExpenses = ref([])
  const recurringLoading = ref(false)
  const recurringLoadError = ref('')

  const filteredRecurring = computed(() => recurringExpenses.value.filter(item => matchesCategory(item)))

  const recurringHeaders = [
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Category', key: 'category_id', align: 'start', sortable: false },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Interval', key: 'interval', align: 'start', sortable: false },
    { title: 'Next Due', key: 'next_due_at', align: 'start' },
    { title: 'Active', key: 'active', align: 'center', sortable: false },
  ]

  async function fetchRecurringExpenses () {
    recurringLoading.value = true
    recurringLoadError.value = ''
    try {
      recurringExpenses.value = await apiJSON(`${API_BASE}/recurring_expenses`) ?? []
    } catch (error) {
      recurringLoadError.value = error.message
    } finally {
      recurringLoading.value = false
    }
  }

  async function toggleRecurringActive (item, active) {
    try {
      await apiJSON(`${API_BASE}/recurring_expenses/active`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: item.id, active }),
      })
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
      await apiJSON(`${API_BASE}/recurring_expenses`, {
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
      closeRecurringDialog()
      fetchRecurringExpenses()
    } catch (error) {
      recurringSubmitError.value = error.message
    } finally {
      recurringSubmitting.value = false
    }
  })

  // -- Categories (sidebar) --

  const categoryDialog = ref(false)
  const editingCategory = ref(null)

  function openCreateCategory () {
    editingCategory.value = null
    categoryDialog.value = true
  }

  function openEditCategory (category) {
    editingCategory.value = category
    categoryDialog.value = true
  }

  async function submitCategory (payload) {
    await (editingCategory.value
      ? updateExpenseCategory({ id: editingCategory.value.id, ...payload })
      : createExpenseCategory(payload))
    await fetchExpenseCategories()
  }

  async function removeCategory (category) {
    if (!confirm(`Delete category "${category.name}"? Categories still used by an expense can't be deleted.`)) return
    try {
      await deleteExpenseCategory(category.id)
      if (selectedCategory.value === category.id) selectedCategory.value = null
      await fetchExpenseCategories()
    } catch (error) {
      expensesError.value = error.message
    }
  }
</script>

<style scoped>
.page-root {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.expenses-card {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.expense-search {
  flex: 0 1 320px;
}

.expenses-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}

.expenses-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}

.expense-row {
  cursor: pointer;
}

.expense-row-action {
  opacity: 0;
  transition: opacity 0.15s ease;
}

.expense-row:hover .expense-row-action,
.expense-row:focus-within .expense-row-action {
  opacity: 1;
}
</style>
