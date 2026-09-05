<template>
  <div class="page-root">
    <v-alert
      v-if="loadError"
      class="mb-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ loadError }}
    </v-alert>

    <v-card class="cp-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-cash-multiple" />
        <span>Currencies &amp; Payment Methods</span>
        <v-spacer />
        <v-text-field
          v-if="tab === 'currencies'"
          v-model="currencySearch"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search currencies"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-text-field
          v-if="tab === 'methods'"
          v-model="methodSearch"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search payment methods"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn
          v-if="tab === 'currencies'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Currency"
          variant="flat"
          @click="openAddCurrency"
        />
        <v-btn
          v-else
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Payment Method"
          variant="flat"
          @click="openAddMethod"
        />
      </v-card-title>
      <v-divider />

      <div class="cp-layout">
        <v-tabs
          v-model="tab"
          class="cp-tabs"
          color="primary"
          direction="vertical"
        >
          <v-tab prepend-icon="mdi-currency-usd" text="Currencies" value="currencies" />
          <v-tab prepend-icon="mdi-credit-card-outline" text="Payment Methods" value="methods" />
        </v-tabs>

        <v-divider vertical />

        <div class="cp-content">
          <!-- ------------------------------------------------------ CURRENCIES -->
          <div v-show="tab === 'currencies'">
            <div class="pa-4">
              <BulkDeleteBar
                :count="selectedCurrencies.length"
                :error="currencySelError"
                :loading="currencyDeleting"
                @clear="resetCurrencySel"
                @confirm="bulkDeleteCurrencies"
              />
              <v-data-table
                v-model:page="currencyPage"
                v-model:selected="selectedCurrencies"
                class="cp-table"
                density="comfortable"
                :headers="currencyHeaders"
                item-value="code"
                :items="currencies"
                :items-per-page="15"
                :items-per-page-options="[15, 25, 50, 100]"
                :loading="currencyLoading"
                :search="currencySearch"
                show-select
                @click:row="openCurrencyRow"
              >
                <template #item.data-table-select="{ item, internalItem, isSelected, toggleSelect }">
                  <v-checkbox-btn
                    :disabled="item.is_default"
                    :model-value="isSelected(internalItem)"
                    @update:model-value="toggleSelect(internalItem)"
                  />
                </template>
                <template #item.is_default="{ item }">
                  <v-chip v-if="item.is_default" color="primary" size="small">Default</v-chip>
                </template>
              </v-data-table>
            </div>
          </div>

          <!-- ------------------------------------------------- PAYMENT METHODS -->
          <div v-show="tab === 'methods'">
            <div class="pa-4">
              <v-data-table
                v-model:page="methodPage"
                class="cp-table"
                density="comfortable"
                :headers="methodHeaders"
                :items="methodRows"
                :items-per-page="15"
                :items-per-page-options="[15, 25, 50, 100]"
                :loading="methodLoading"
                :search="methodSearch"
                @click:row="openMethodRow"
              >
                <template #item.color="{ item }">
                  <div v-if="item.color" class="d-flex align-center ga-2">
                    <div class="swatch" :style="{ backgroundColor: item.color }" />
                    <span class="text-caption">{{ item.color }}</span>
                  </div>
                  <span v-else class="text-disabled">—</span>
                </template>
                <template #item.reserved="{ item }">
                  <v-chip v-if="item.reserved" size="small" variant="tonal">System</v-chip>
                </template>
              </v-data-table>
            </div>
          </div>
        </div>
      </div>
    </v-card>

    <!-- add / edit currency -->
    <v-dialog v-model="currencyDialog" max-width="480">
      <v-card class="px-4">
        <v-card-title>{{ editingCode ? 'Edit Currency' : 'Add a New Currency' }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="currencyForm.code"
            density="compact"
            :disabled="!!editingCode"
            hint="e.g. USD, EUR"
            label="Code"
          />
          <v-text-field
            v-model="currencyForm.name"
            density="compact"
            label="Name"
          />
          <v-text-field
            v-model="currencyForm.symbol"
            density="compact"
            label="Symbol"
          />
          <v-text-field
            v-model.number="currencyForm.valueInDefaultCurrency"
            density="compact"
            label="Value in Default Currency"
            type="number"
          />

          <v-alert
            v-if="currencyError"
            class="mb-2 mt-4"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ currencyError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-btn
            v-if="editingCode"
            color="error"
            :disabled="editingIsDefault"
            prepend-icon="mdi-delete"
            text="Delete"
            variant="text"
            @click="deleteCurrentCurrency"
          />
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="currencyDialog = false" />
          <v-btn
            color="primary"
            :loading="currencySaving"
            :text="editingCode ? 'Save' : 'Add Currency'"
            variant="flat"
            @click="saveCurrency"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- delete currency -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title>Delete Currency</v-card-title>
        <v-card-text>
          Delete <strong>{{ deleteTarget?.name }}</strong>? This cannot be undone.
          <v-alert
            v-if="deleteError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ deleteError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="deleteDialog = false" />
          <v-btn
            color="error"
            :loading="deleting"
            text="Delete"
            variant="flat"
            @click="confirmDeleteCurrency"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- add / edit payment method -->
    <v-dialog v-model="methodDialog" max-width="420">
      <v-card>
        <v-card-title class="d-flex align-center ga-2 px-4 py-3 text-h6">
          <v-icon icon="mdi-credit-card-outline" size="22" />
          <span>{{ editingMethodId ? 'Edit Payment Method' : 'Add Payment Method' }}</span>
          <v-spacer />
          <v-btn icon="mdi-close" size="small" variant="text" @click="methodDialog = false" />
        </v-card-title>
        <v-divider />
        <v-card-text class="px-4 pt-3">
          <v-text-field
            v-model="methodForm.name"
            density="compact"
            hide-details="auto"
            label="Name"
            placeholder="Cash / Card / Transfer"
            variant="outlined"
          />
          <CurrencySelect v-model="methodForm.currencyCode" class="mt-3" />
          <v-text-field
            v-model.number="methodForm.sortOrder"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Order"
            type="number"
            variant="outlined"
          />
          <v-text-field
            v-model="methodForm.color"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Accent color (optional)"
            placeholder="#4CAF50"
            variant="outlined"
          >
            <template #append-inner>
              <div class="swatch" :style="{ backgroundColor: isValidHex(methodForm.color) ? methodForm.color : 'transparent' }" />
            </template>
          </v-text-field>
          <v-color-picker
            v-model="methodForm.color"
            class="method-color-picker mt-3"
            elevation="0"
            hide-inputs
            mode="hex"
            width="100%"
          />
          <v-alert
            v-if="methodError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ methodError }}</v-alert>
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="methodDialog = false" />
          <v-btn
            color="primary"
            :loading="methodSaving"
            :text="editingMethodId ? 'Save' : 'Add'"
            variant="flat"
            @click="saveMethod"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import CurrencySelect from '@/components/CurrencySelect.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useCashboxAccounts } from '@/composables/useCashboxAccounts'
  import { useCurrencies } from '@/composables/useCurrencies'

  const {
    currencies,
    fetchCurrencies,
    createCurrency,
    updateCurrency,
    deleteCurrency,
  } = useCurrencies()

  const {
    selected: selectedCurrencies,
    deleting: currencyDeleting,
    error: currencySelError,
    reset: resetCurrencySel,
    run: runCurrencyBulk,
  } = useBulkDelete()

  async function bulkDeleteCurrencies () {
    try {
      await runCurrencyBulk('/currencies/bulk_delete', { codes: selectedCurrencies.value })
      resetCurrencySel()
      reloadCurrencies()
    } catch { /* error shown in the bar */ }
  }
  const {
    cashboxAccounts,
    fetchCashboxAccounts,
    createCashboxAccount,
    updateCashboxAccount,
  } = useCashboxAccounts()

  const tab = ref('currencies')
  const loadError = ref('')
  const currencyLoading = ref(false)
  const methodLoading = ref(false)
  const currencySearch = ref('')
  const currencyPage = ref(1)
  const methodSearch = ref('')
  const methodPage = ref(1)

  // Reset to the first page whenever a search narrows the list.
  watch(currencySearch, () => { currencyPage.value = 1 })
  watch(methodSearch, () => { methodPage.value = 1 })

  const currencyHeaders = [
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Symbol', key: 'symbol', align: 'start' },
    { title: 'Value in Default Currency', key: 'value_in_default_currency', align: 'end' },
    { title: '', key: 'is_default', align: 'center', sortable: false },
  ]

  const methodHeaders = [
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Currency', key: 'currency_code', align: 'start' },
    { title: 'Color', key: 'color', align: 'start' },
    { title: 'Order', key: 'sort_order', align: 'end', width: 90 },
    { title: '', key: 'reserved', align: 'center', sortable: false },
  ]

  // "On Account" / "On Coupon" are internal bookkeeping tenders in kashi-pos
  // and can't be edited there, so flag them read-only here too.
  function isReserved (name) {
    const n = String(name ?? '').trim().toLowerCase()
    return n === 'on account' || n === 'on coupon'
  }

  const methodRows = computed(() =>
    [...cashboxAccounts.value]
      .map(m => ({ ...m, reserved: isReserved(m.name) }))
      .toSorted((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0)),
  )

  function isValidHex (value) {
    return /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(value ?? '')
  }

  // v-color-picker hands back #RRGGBBAA once the alpha slider is touched; keep
  // the plain #RRGGBB the API expects.
  function normalizeHex (value) {
    return typeof value === 'string' && value.length === 9 ? value.slice(0, 7) : value
  }

  onMounted(() => {
    reloadCurrencies()
    reloadMethods()
  })

  async function reloadCurrencies () {
    currencyLoading.value = true
    loadError.value = ''
    try {
      await fetchCurrencies(true)
    } catch (error) {
      loadError.value = error.message
    } finally {
      currencyLoading.value = false
    }
  }

  async function reloadMethods () {
    methodLoading.value = true
    loadError.value = ''
    try {
      await fetchCashboxAccounts(true)
    } catch (error) {
      loadError.value = error.message
    } finally {
      methodLoading.value = false
    }
  }

  // ---- currency add / edit -------------------------------------------------
  const currencyDialog = ref(false)
  const currencySaving = ref(false)
  const currencyError = ref('')
  const editingCode = ref(null)
  const currencyForm = reactive({ code: '', name: '', symbol: '', valueInDefaultCurrency: 1 })

  function openAddCurrency () {
    editingCode.value = null
    currencyError.value = ''
    Object.assign(currencyForm, { code: '', name: '', symbol: '', valueInDefaultCurrency: 1 })
    currencyDialog.value = true
  }

  const editingIsDefault = ref(false)

  function openEditCurrency (item) {
    editingCode.value = item.code
    editingIsDefault.value = !!item.is_default
    currencyError.value = ''
    Object.assign(currencyForm, {
      code: item.code,
      name: item.name,
      symbol: item.symbol,
      valueInDefaultCurrency: item.value_in_default_currency,
    })
    currencyDialog.value = true
  }

  function openCurrencyRow (_event, { item }) {
    openEditCurrency(item)
  }

  // Delete from inside the edit dialog: hand off to the existing confirm dialog.
  function deleteCurrentCurrency () {
    const item = currencies.value.find(c => c.code === editingCode.value)
    if (!item) return
    currencyDialog.value = false
    openDeleteCurrency(item)
  }

  async function saveCurrency () {
    if (!editingCode.value && (currencyForm.code ?? '').trim().length < 2) {
      currencyError.value = 'Code is required.'
      return
    }
    if ((currencyForm.name ?? '').trim().length < 2) {
      currencyError.value = 'Name needs to be at least 2 characters.'
      return
    }
    if (!(currencyForm.symbol ?? '').trim()) {
      currencyError.value = 'Symbol is required.'
      return
    }
    if (!(Number(currencyForm.valueInDefaultCurrency) > 0)) {
      currencyError.value = 'Value in default currency must be greater than 0.'
      return
    }
    currencySaving.value = true
    currencyError.value = ''
    try {
      const payload = {
        code: currencyForm.code.trim(),
        name: currencyForm.name.trim(),
        symbol: currencyForm.symbol.trim(),
        value_in_default_currency: Number(currencyForm.valueInDefaultCurrency),
      }
      await (editingCode.value ? updateCurrency(payload) : createCurrency(payload))
      currencyDialog.value = false
      reloadCurrencies()
    } catch (error) {
      currencyError.value = error.message
    } finally {
      currencySaving.value = false
    }
  }

  // ---- currency delete ---------------------------------------------------
  const deleteDialog = ref(false)
  const deleteTarget = ref(null)
  const deleting = ref(false)
  const deleteError = ref('')

  function openDeleteCurrency (item) {
    deleteTarget.value = item
    deleteError.value = ''
    deleteDialog.value = true
  }

  async function confirmDeleteCurrency () {
    deleting.value = true
    deleteError.value = ''
    try {
      await deleteCurrency(deleteTarget.value.code)
      deleteDialog.value = false
      reloadCurrencies()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }

  // ---- payment method add / edit ---------------------------------------
  const methodDialog = ref(false)
  const methodSaving = ref(false)
  const methodError = ref('')
  const editingMethodId = ref(null)
  const methodForm = reactive({ name: '', currencyCode: 'USD', sortOrder: 0, color: '' })

  function openAddMethod () {
    editingMethodId.value = null
    methodError.value = ''
    const nextOrder = methodRows.value.reduce((max, m) => Math.max(max, m.sort_order ?? 0), 0) + 1
    Object.assign(methodForm, { name: '', currencyCode: 'USD', sortOrder: nextOrder, color: '' })
    methodDialog.value = true
  }

  function openEditMethod (item) {
    if (item.reserved) return
    editingMethodId.value = item.id
    methodError.value = ''
    Object.assign(methodForm, {
      name: item.name,
      currencyCode: item.currency_code || 'USD',
      sortOrder: item.sort_order ?? 0,
      color: item.color || '',
    })
    methodDialog.value = true
  }

  function openMethodRow (_event, { item }) {
    openEditMethod(item)
  }

  async function saveMethod () {
    if (!(methodForm.name ?? '').trim()) {
      methodError.value = 'Name is required.'
      return
    }
    const color = normalizeHex(methodForm.color)
    if (color && !isValidHex(color)) {
      methodError.value = 'Accent color must be a valid hex value.'
      return
    }
    methodSaving.value = true
    methodError.value = ''
    try {
      const payload = {
        name: methodForm.name.trim(),
        currency_code: methodForm.currencyCode || 'USD',
        sort_order: Number(methodForm.sortOrder) || 0,
        color: color || '',
      }
      await (editingMethodId.value ? updateCashboxAccount({ id: editingMethodId.value, ...payload }) : createCashboxAccount(payload));
      methodDialog.value = false
      reloadMethods()
    } catch (error) {
      methodError.value = error.message
    } finally {
      methodSaving.value = false
    }
  }
</script>

<style scoped>
/* Keep intrinsic height inside the layout's flex-column scroll wrapper so the
   scroll lives inside the tab content, not the whole page. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.cp-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.cp-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.cp-tabs {
  flex: 0 0 180px;
}
.cp-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.table-search {
  flex: 0 1 320px;
}
.swatch {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 1px solid rgba(128, 128, 128, 0.6);
}
.method-color-picker {
  max-width: 300px;
  margin-inline: auto;
}
:deep(.cp-table tbody tr) {
  cursor: pointer;
}
</style>
