<!--
  Master/detail manager shared by pages/price-lists.vue and
  pages/discount-lists.vue. They are the same entity: a list of lists, each
  with per-product line items and a branch assignment. The only real
  difference is one value column — `price` (cents) vs `discount` (percent) —
  captured in `cfg` below and keyed off the `kind` prop.

  Line items are paged server-side through ServerSideTable; nothing here
  holds the whole item set in memory.
-->
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

    <v-card class="vlm-card" flat>
      <v-card-title class="page-heading d-flex align-center ga-3 px-4 py-3">
        <v-icon :icon="cfg.icon" />
        <span>{{ cfg.title }}</span>
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          :text="`Add ${cfg.singularTitle}`"
          variant="flat"
          @click="openCreateList"
        />
      </v-card-title>
      <v-divider />

      <div class="vlm-layout">
        <PageSidebar v-model="selectedListId" :items="listItems" :width="240">
          <template #append="{ item }">
            <v-chip v-if="item.is_default" color="primary" size="x-small" text="Default" />
            <v-icon
              v-else-if="!item.is_active"
              color="grey"
              icon="mdi-close-circle"
              size="16"
            />
          </template>
        </PageSidebar>

        <v-divider vertical />

        <div class="vlm-content">
          <div v-if="!selectedList" class="pa-6 text-medium-emphasis">
            Select a {{ cfg.singular }} on the left, or add one.
          </div>

          <template v-else>
            <div class="d-flex align-center flex-wrap ga-2 pa-4 pb-2">
              <span class="text-h6">{{ selectedList.name }}</span>
              <v-chip v-if="selectedList.is_default" color="primary" size="small" text="Default" />
              <v-chip
                :color="selectedList.is_active ? 'success' : 'grey'"
                size="small"
                :text="selectedList.is_active ? 'Active' : 'Inactive'"
              />
              <span class="text-medium-emphasis text-body-2">
                {{ formatDate(selectedList.valid_from) }} – {{ formatDate(selectedList.valid_to) }}
              </span>
              <v-spacer />
              <v-btn
                prepend-icon="mdi-store-outline"
                size="small"
                text="Branches"
                variant="tonal"
                @click="branchesDialog = true"
              />
              <v-btn icon="mdi-pencil" size="small" variant="text" @click="openEditList" />
              <v-btn
                color="error"
                icon="mdi-delete"
                size="small"
                variant="text"
                @click="deleteListDialog = true"
              />
            </div>

            <v-divider />

            <div class="d-flex align-end ga-3 pa-4">
              <v-autocomplete
                v-model="addProductId"
                class="flex-grow-1"
                clearable
                density="compact"
                hide-details
                :item-title="p => `${p.code} — ${p.name}`"
                item-value="id"
                :items="productOptions"
                label="Product"
                :loading="productSearchLoading"
                no-filter
                variant="outlined"
                @update:search="onProductSearch"
              />
              <v-text-field
                v-model="addValue"
                density="compact"
                hide-details
                :label="cfg.valueLabel"
                :prefix="cfg.prefix"
                style="max-width: 140px;"
                :suffix="cfg.suffix"
                type="number"
                variant="outlined"
                @keyup.enter="addItem"
              />
              <v-btn
                color="primary"
                :disabled="!addProductId || addValue === ''"
                :loading="addingItem"
                text="Add"
                variant="flat"
                @click="addItem"
              />
            </div>

            <v-text-field
              v-model="itemSearch"
              class="mx-4"
              clearable
              density="compact"
              hide-details
              label="Search products in this list"
              prepend-inner-icon="mdi-magnify"
              variant="outlined"
            />

            <v-alert
              v-if="itemError"
              class="mx-4 mb-2 mt-2"
              density="compact"
              type="error"
              variant="tonal"
            >
              {{ itemError }}
            </v-alert>

            <BulkDeleteBar
              :count="selectedItems.length"
              :error="itemSelError"
              :loading="itemDeleting"
              @clear="resetItemSel"
              @confirm="bulkDeleteItems"
            />

            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="itemsURL(selectedList.id)"
              class="mt-2"
              density="comfortable"
              :external-search="itemSearch ?? ''"
              flush
              :headers="itemHeaders"
              item-value="product_id"
              root-key="items"
              selectable
              :show-search-icon="false"
              @row-click="openEditItem"
              @update:selected="v => (selectedItems = v)"
            >
              <template #[`item.${cfg.valueKey}`]="{ item }">
                {{ cfg.display(item[cfg.valueKey]) }}
              </template>
            </ServerSideTable>
          </template>
        </div>
      </div>
    </v-card>

    <!-- create / edit list -->
    <v-dialog v-model="listDialog" max-width="520">
      <v-card>
        <v-card-title>{{ editingList ? `Edit ${cfg.singularTitle}` : `Add ${cfg.singularTitle}` }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="listForm.name" density="compact" label="Name" variant="outlined" />
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model="listForm.validFrom"
                density="compact"
                label="Valid From"
                type="date"
                variant="outlined"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model="listForm.validTo"
                density="compact"
                label="Valid To"
                type="date"
                variant="outlined"
              />
            </v-col>
          </v-row>
          <div class="d-flex ga-6">
            <v-switch
              v-model="listForm.isActive"
              color="primary"
              density="compact"
              hide-details
              label="Active"
            />
            <v-switch
              v-model="listForm.isDefault"
              color="primary"
              density="compact"
              hide-details
              label="Default"
            />
          </div>
          <v-alert
            v-if="listError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ listError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="listDialog = false" />
          <v-btn
            color="primary"
            :loading="savingList"
            text="Save"
            variant="flat"
            @click="saveList"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- edit item value -->
    <v-dialog v-model="editItemDialog" max-width="420">
      <v-card>
        <v-card-title>{{ editItemTarget?.product_code }} — {{ editItemTarget?.product_name }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="editItemValue"
            autofocus
            density="compact"
            hide-details
            :label="cfg.valueLabel"
            :prefix="cfg.prefix"
            :suffix="cfg.suffix"
            type="number"
            variant="outlined"
            @keyup.enter="saveEditItem"
          />
        </v-card-text>
        <v-card-actions>
          <v-btn
            color="error"
            prepend-icon="mdi-delete"
            text="Remove"
            variant="text"
            @click="deleteCurrentItem"
          />
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="editItemDialog = false" />
          <v-btn
            color="primary"
            :loading="savingItem"
            text="Save"
            variant="flat"
            @click="saveEditItem"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- delete list -->
    <v-dialog v-model="deleteListDialog" max-width="420">
      <v-card>
        <v-card-title>Delete {{ cfg.singular }}</v-card-title>
        <v-card-text>
          Delete <strong>{{ selectedList?.name }}</strong> and all its entries? This can't be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="deleteListDialog = false" />
          <v-btn
            color="error"
            :loading="deletingList"
            text="Delete"
            variant="flat"
            @click="doDeleteList"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <BranchAssignDialog
      v-model="branchesDialog"
      :kind="cfg.singular"
      :load-fn="() => fetchBranchIds(selectedList.id)"
      :save-fn="ids => setBranches(selectedList.id, ids)"
      :title="`Branches using “${selectedList?.name}”`"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import BranchAssignDialog from '@/components/BranchAssignDialog.vue'
  import PageSidebar from '@/components/PageSidebar.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { authFetch } from '@/composables/useApi'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useDiscountLists } from '@/composables/useDiscountLists'
  import { usePriceLists } from '@/composables/usePriceLists'
  import { API_BASE } from '@/config'
  import { formatMoney } from '@/utils/money'

  const props = defineProps({
    kind: { type: String, required: true, validator: v => v === 'price' || v === 'discount' },
  })

  const CONFIGS = {
    price: {
      icon: 'mdi-format-list-bulleted',
      title: 'Price Lists',
      singular: 'price list',
      singularTitle: 'Price List',
      valueKey: 'price',
      valueLabel: 'Price',
      prefix: '$',
      suffix: undefined,
      parse: v => Math.round(Number.parseFloat(v) * 100),
      toInput: v => (v / 100).toString(),
      display: v => `$${formatMoney(v)}`,
    },
    discount: {
      icon: 'mdi-sale',
      title: 'Discount Lists',
      singular: 'discount list',
      singularTitle: 'Discount List',
      valueKey: 'discount',
      valueLabel: 'Discount',
      prefix: undefined,
      suffix: '%',
      parse: v => Math.round(Number.parseFloat(v)),
      toInput: String,
      display: v => `${v}%`,
    },
  }
  const cfg = computed(() => CONFIGS[props.kind])

  const store = props.kind === 'price' ? usePriceLists() : useDiscountLists()
  const {
    list: lists, fetchLists, createList, updateList, deleteList,
    itemsURL, saveItem, deleteItem, fetchBranchIds, setBranches,
  } = store

  const itemHeaders = computed(() => [
    { title: 'Code', key: 'product_code', align: 'start', width: 140 },
    { title: 'Product', key: 'product_name', align: 'start' },
    { title: cfg.value.valueLabel, key: cfg.value.valueKey, align: 'end', width: 140 },
  ])

  const loadError = ref('')
  const selectedListId = ref(null)
  const selectedList = computed(() => lists.value.find(l => l.id === selectedListId.value) ?? null)
  const listItems = computed(() => lists.value.map(l => ({ value: l.id, title: l.name, raw: l })))
  const tableRef = ref(null)
  const itemSearch = ref('')
  const itemError = ref('')

  function reloadItems () {
    tableRef.value?.reload()
  }

  // ---- item bulk delete ------------------------------------------------
  const {
    selected: selectedItems,
    deleting: itemDeleting,
    error: itemSelError,
    reset: resetItemSel,
    run: runItemBulk,
  } = useBulkDelete()

  async function bulkDeleteItems () {
    if (!selectedList.value) return
    try {
      await runItemBulk(`/${props.kind === 'price' ? 'price_lists' : 'discount_lists'}/${selectedList.value.id}/items/bulk_delete`, {
        product_ids: selectedItems.value,
      })
      resetItemSel()
      reloadItems()
    } catch { /* error shown in the bar */ }
  }

  watch(selectedListId, () => {
    resetItemSel()
    itemSearch.value = ''
    itemError.value = ''
  })

  onMounted(async () => {
    try {
      await fetchLists(true)
      if (selectedListId.value == null && lists.value.length > 0) {
        selectedListId.value = lists.value[0].id
      }
    } catch (error) {
      loadError.value = error.message
    }
  })

  // ---- product search -------------------------------------------------
  const productOptions = ref([])
  const productSearchLoading = ref(false)
  const addProductId = ref(null)
  const addValue = ref('')
  const addingItem = ref(false)
  let productSearchTimer

  function onProductSearch (query) {
    clearTimeout(productSearchTimer)
    productSearchTimer = setTimeout(() => searchProducts(query), 300)
  }

  async function searchProducts (query) {
    productSearchLoading.value = true
    try {
      const url = new URL(`${API_BASE}/products`)
      url.searchParams.set('page_size', '20')
      url.searchParams.set('page_id', '0')
      if (query && query.trim()) url.searchParams.set('search', query.trim())
      const res = await authFetch(url)
      const data = await res.json().catch(() => ({}))
      productOptions.value = data.products ?? []
    } catch {
      productOptions.value = []
    } finally {
      productSearchLoading.value = false
    }
  }

  onMounted(() => searchProducts(''))

  async function addItem () {
    if (!addProductId.value || addValue.value === '') return
    const parsed = cfg.value.parse(addValue.value)
    if (!Number.isFinite(parsed) || parsed < 0) {
      itemError.value = `Enter a valid ${cfg.value.valueLabel.toLowerCase()}.`
      return
    }
    addingItem.value = true
    itemError.value = ''
    try {
      await saveItem(selectedList.value.id, addProductId.value, parsed)
      addProductId.value = null
      addValue.value = ''
      reloadItems()
    } catch (error) {
      itemError.value = error.message
    } finally {
      addingItem.value = false
    }
  }

  // ---- edit / delete item -------------------------------------------------
  const editItemDialog = ref(false)
  const editItemTarget = ref(null)
  const editItemValue = ref('')
  const savingItem = ref(false)

  function openEditItem (item) {
    editItemTarget.value = item
    editItemValue.value = cfg.value.toInput(item[cfg.value.valueKey])
    editItemDialog.value = true
  }

  function deleteCurrentItem () {
    const item = editItemTarget.value
    editItemDialog.value = false
    if (item) removeItem(item)
  }

  async function saveEditItem () {
    const parsed = cfg.value.parse(editItemValue.value)
    if (!Number.isFinite(parsed) || parsed < 0) return
    savingItem.value = true
    try {
      await saveItem(selectedList.value.id, editItemTarget.value.product_id, parsed)
      editItemDialog.value = false
      reloadItems()
    } catch (error) {
      itemError.value = error.message
    } finally {
      savingItem.value = false
    }
  }

  async function removeItem (item) {
    try {
      await deleteItem(selectedList.value.id, item.product_id)
      reloadItems()
    } catch (error) {
      itemError.value = error.message
    }
  }

  // ---- create / edit / delete list ------------------------------------
  const listDialog = ref(false)
  const editingList = ref(false)
  const savingList = ref(false)
  const listError = ref('')
  const listForm = reactive({ name: '', validFrom: '', validTo: '', isActive: true, isDefault: false })

  function openCreateList () {
    editingList.value = false
    Object.assign(listForm, { name: '', validFrom: '', validTo: '', isActive: true, isDefault: false })
    listError.value = ''
    listDialog.value = true
  }

  function openEditList () {
    editingList.value = true
    Object.assign(listForm, {
      name: selectedList.value.name,
      validFrom: selectedList.value.valid_from?.slice(0, 10) ?? '',
      validTo: selectedList.value.valid_to?.slice(0, 10) ?? '',
      isActive: selectedList.value.is_active,
      isDefault: selectedList.value.is_default,
    })
    listError.value = ''
    listDialog.value = true
  }

  async function saveList () {
    if (!listForm.name.trim() || !listForm.validFrom || !listForm.validTo) {
      listError.value = 'Name and both dates are required.'
      return
    }
    savingList.value = true
    listError.value = ''
    const payload = {
      name: listForm.name.trim(),
      is_active: listForm.isActive,
      is_default: listForm.isDefault,
      valid_from: new Date(listForm.validFrom).toISOString(),
      valid_to: new Date(listForm.validTo).toISOString(),
    }
    try {
      if (editingList.value) {
        await updateList({ id: selectedList.value.id, ...payload })
      } else {
        const created = await createList(payload)
        if (created?.id) selectedListId.value = created.id
      }
      listDialog.value = false
    } catch (error) {
      listError.value = error.message
    } finally {
      savingList.value = false
    }
  }

  const deleteListDialog = ref(false)
  const deletingList = ref(false)

  async function doDeleteList () {
    deletingList.value = true
    try {
      await deleteList(selectedList.value.id)
      deleteListDialog.value = false
      selectedListId.value = lists.value.length > 0 ? lists.value[0].id : null
    } catch (error) {
      loadError.value = error.message
    } finally {
      deletingList.value = false
    }
  }

  const branchesDialog = ref(false)

  function formatDate (value) {
    return value ? new Date(value).toLocaleDateString() : '—'
  }
</script>

<style scoped>
.page-root, .vlm-card {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}
.vlm-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.vlm-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
</style>
