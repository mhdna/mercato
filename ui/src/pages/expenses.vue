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
          <v-row dense>
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
              <v-text-field
                v-model="currencyCode.value.value"
                density="compact"
                :error-messages="currencyCode.errorMessage.value"
                label="Currency"
              />
            </v-col>
          </v-row>

          <v-alert v-if="expenseSubmitError" class="mb-4" type="error" variant="tonal">
            {{ expenseSubmitError }}
          </v-alert>

          <v-card-actions class="px-0">
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
          <v-row dense>
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
              <v-text-field
                v-model="rCurrencyCode.value.value"
                density="compact"
                :error-messages="rCurrencyCode.errorMessage.value"
                label="Currency"
              />
            </v-col>
          </v-row>
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

  <!-- Manage the category list -->
  <v-dialog v-model="categoriesDialog" max-width="520" scrollable>
    <v-card>
      <v-card-title class="d-flex align-center">
        Expense Categories
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          size="small"
          text="Add"
          @click="openCreateCategory"
        />
      </v-card-title>
      <v-divider />
      <v-card-text style="max-height: 62vh">
        <v-list class="py-0" lines="one">
          <v-list-item
            v-for="category in sortedCategories"
            :key="category.id"
            class="px-1"
            :class="{ 'text-medium-emphasis': !category.is_active }"
          >
            <template #prepend>
              <v-avatar
                class="me-3"
                :color="categoryColor(category)"
                rounded="lg"
                size="38"
                variant="tonal"
              >
                <v-icon :color="categoryColor(category)" :icon="categoryIcon(category)" size="20" />
              </v-avatar>
            </template>
            <v-list-item-title class="font-weight-medium">{{ category.name }}</v-list-item-title>
            <v-list-item-subtitle v-if="!category.is_active">Inactive</v-list-item-subtitle>
            <template #append>
              <v-btn
                density="comfortable"
                icon="mdi-pencil"
                size="small"
                variant="text"
                @click="openEditCategory(category)"
              />
              <v-btn
                density="comfortable"
                icon="mdi-delete"
                size="small"
                variant="text"
                @click="removeCategory(category)"
              />
            </template>
          </v-list-item>
        </v-list>
        <div v-if="expenseCategories.length === 0" class="text-medium-emphasis text-center pa-8">
          No categories yet — add your first one.
        </div>
        <v-alert v-if="categoryListError" class="mt-2" type="error" variant="tonal">{{ categoryListError }}</v-alert>
      </v-card-text>
      <v-divider />
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="categoriesDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Add / edit a single category (opens above the manager dialog) -->
  <v-dialog v-model="categoryDialog" max-width="540">
    <v-card class="px-4">
      <v-card-title>{{ editingCategoryId ? 'Edit Category' : 'Add a New Category' }}</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitCategory">
          <div class="d-flex align-center ga-3 mb-4">
            <v-avatar :color="pickedColor" rounded="lg" size="48" variant="tonal">
              <v-icon :color="pickedColor" :icon="pickedIcon" size="26" />
            </v-avatar>
            <v-text-field
              v-model="categoryName.value.value"
              density="compact"
              :error-messages="categoryName.errorMessage.value"
              hide-details="auto"
              label="Name"
            />
          </div>

          <div class="text-caption text-medium-emphasis mb-1">Icon</div>
          <div class="d-flex flex-wrap ga-1 mb-4">
            <v-btn
              v-for="icon in EXPENSE_CATEGORY_ICONS"
              :key="icon"
              :active="pickedIcon === icon"
              :color="pickedIcon === icon ? pickedColor : undefined"
              density="comfortable"
              :icon="icon"
              size="small"
              variant="tonal"
              @click="categoryIconField.value.value = icon"
            />
          </div>

          <div class="text-caption text-medium-emphasis mb-1">Colour</div>
          <div class="d-flex flex-wrap ga-2 mb-4">
            <v-btn
              v-for="color in EXPENSE_CATEGORY_COLORS"
              :key="color"
              :color="color"
              :icon="pickedColor === color ? 'mdi-check' : 'mdi-circle'"
              size="x-small"
              @click="categoryColorField.value.value = color"
            />
          </div>

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
          <v-btn prepend-icon="mdi-tag-multiple" text="Categories" variant="tonal" @click="categoriesDialog = true" />
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
        </div>
      </v-card-title>
      <v-divider />

      <v-window v-model="tab">
        <!-- Mine -->
        <v-window-item value="own">
          <v-alert v-if="expensesError" class="ma-4" type="error" variant="tonal">{{ expensesError }}</v-alert>

          <v-list v-if="filteredExpenses.length > 0" class="py-0" lines="two">
            <template v-for="(item, i) in filteredExpenses" :key="item.id">
              <v-list-item class="expense-row px-4 py-2" @click="openEditExpense(item)">
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
                <v-list-item-subtitle>{{ prettyDate(item.created_at) }} · {{ categoryLabel(item.category_id) }}</v-list-item-subtitle>
                <template #append>
                  <span class="text-error font-weight-medium">−{{ money(item.amount) }} {{ item.currency_code }}</span>
                  <v-btn
                    class="expense-row-action ms-1"
                    density="comfortable"
                    icon="mdi-delete"
                    size="small"
                    variant="text"
                    @click.stop="removeExpense(item)"
                  />
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
            <v-data-table density="compact" :headers="recurringHeaders" :items="recurringExpenses" :loading="recurringLoading">
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
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { useField, useForm } from 'vee-validate'
  import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
  import CategoryChip from '@/components/CategoryChip.vue'
  import { authFetch } from '@/composables/useApi'
  import { useBranches } from '@/composables/useBranches'
  import { useExpenseCategories } from '@/composables/useExpenseCategories'
  import { API_BASE } from '@/config'
  import {
    categoryColor,
    categoryIcon,
    DEFAULT_EXPENSE_CATEGORY_COLOR,
    DEFAULT_EXPENSE_CATEGORY_ICON,
    EXPENSE_CATEGORY_COLORS,
    EXPENSE_CATEGORY_ICONS,
  } from '@/data/expenseCategoryIcons'

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

  const activeExpenseCategories = computed(() => expenseCategories.value.filter(c => c.is_active))

  const sortedCategories = computed(() => expenseCategories.value.toSorted((a, b) => {
    if (a.is_active !== b.is_active) return a.is_active ? -1 : 1
    return a.name.localeCompare(b.name)
  }))

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

  const filteredExpenses = computed(() => {
    const query = search.value.trim().toLowerCase()
    if (!query) return expenses.value
    return expenses.value.filter(item => [
      item.description,
      item.currency_code,
      categoryLabel(item.category_id),
    ].some(value => String(value ?? '').toLowerCase().includes(query)))
  })

  async function fetchExpenses (append = false) {
    expensesLoading.value = true
    expensesError.value = ''
    try {
      const offset = append ? expenses.value.length : 0
      const data = await apiJSON(`${API_BASE}/expenses?page_size=${PAGE_SIZE}&page_id=${offset}`)
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

  const filteredBranchExpenses = computed(() => {
    const query = search.value.trim().toLowerCase()
    if (!query) return branchExpenses.value
    return branchExpenses.value.filter(item => [
      item.description,
      item.currency_code,
      branchName(item.branch_id),
      categoryLabel(item.category_id),
    ].some(value => String(value ?? '').toLowerCase().includes(query)))
  })

  async function fetchBranchExpenses (append = false) {
    branchLoading.value = true
    branchError.value = ''
    try {
      const offset = append ? branchExpenses.value.length : 0
      const url = new URL(`${API_BASE}/branch_expenses`, window.location.origin)
      url.searchParams.set('page_size', PAGE_SIZE)
      url.searchParams.set('page_id', offset)
      if (branchFilter.value) url.searchParams.set('branch_id', branchFilter.value)
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

  watch(tab, value => {
    if (value === 'branch' && !branchLoaded) fetchBranchExpenses()
    if (value === 'recurring' && recurringExpenses.value.length === 0) fetchRecurringExpenses()
    nextTick(setupInfiniteScroll)
  })

  // -- Recurring expenses --

  const recurringExpenses = ref([])
  const recurringLoading = ref(false)
  const recurringLoadError = ref('')

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

  // -- Categories --

  const categoriesDialog = ref(false)
  const categoryListError = ref('')
  const categoryDialog = ref(false)
  const categorySubmitting = ref(false)
  const categorySubmitError = ref('')
  const editingCategoryId = ref(null)

  const {
    handleSubmit: handleCategorySubmit,
    handleReset: handleCategoryReset,
    setValues: setCategoryValues,
  } = useForm({
    validationSchema: {
      categoryName (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
    },
  })

  const categoryName = useField('categoryName')
  const categoryIsActive = useField('categoryIsActive')
  const categoryIconField = useField('categoryIcon')
  const categoryColorField = useField('categoryColor')

  const pickedIcon = computed(() => categoryIconField.value.value || DEFAULT_EXPENSE_CATEGORY_ICON)
  const pickedColor = computed(() => categoryColorField.value.value || DEFAULT_EXPENSE_CATEGORY_COLOR)

  function openCreateCategory () {
    editingCategoryId.value = null
    handleCategoryReset()
    categoryIsActive.value.value = true
    categoryIconField.value.value = DEFAULT_EXPENSE_CATEGORY_ICON
    categoryColorField.value.value = DEFAULT_EXPENSE_CATEGORY_COLOR
    categoryDialog.value = true
  }

  function openEditCategory (category) {
    editingCategoryId.value = category.id
    setCategoryValues({
      categoryName: category.name,
      categoryIsActive: category.is_active,
      categoryIcon: categoryIcon(category),
      categoryColor: categoryColor(category),
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
      const payload = {
        name: values.categoryName,
        is_active: !!values.categoryIsActive,
        icon: values.categoryIcon || DEFAULT_EXPENSE_CATEGORY_ICON,
        color: values.categoryColor || DEFAULT_EXPENSE_CATEGORY_COLOR,
      }
      await (editingCategoryId.value
        ? updateExpenseCategory({ id: editingCategoryId.value, ...payload })
        : createExpenseCategory(payload))
      closeCategoryDialog()
    } catch (error) {
      categorySubmitError.value = error.message
    } finally {
      categorySubmitting.value = false
    }
  })

  async function removeCategory (category) {
    if (!confirm(`Delete category "${category.name}"? Categories still used by an expense can't be deleted.`)) return
    categoryListError.value = ''
    try {
      await deleteExpenseCategory(category.id)
    } catch (error) {
      categoryListError.value = error.message
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
