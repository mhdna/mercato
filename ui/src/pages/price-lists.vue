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

    <v-card class="pl-card" flat>
      <v-card-title class="page-heading d-flex align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-format-list-bulleted" />
        <span>Price Lists</span>
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Price List"
          variant="flat"
          @click="openCreateList"
        />
      </v-card-title>
      <v-divider />

      <div class="pl-layout">
        <div class="pl-sidebar">
          <v-list
            v-model:selected="selection"
            class="py-0"
            density="compact"
            mandatory
            nav
            select-strategy="single-leaf"
          >
            <v-list-item
              v-for="pl in priceLists"
              :key="pl.id"
              :value="pl.id"
            >
              <v-list-item-title>{{ pl.name }}</v-list-item-title>
              <template #append>
                <v-chip v-if="pl.is_default" color="primary" size="x-small" text="Default" />
                <v-icon
                  v-else-if="!pl.is_active"
                  color="grey"
                  icon="mdi-close-circle"
                  size="16"
                />
              </template>
            </v-list-item>
          </v-list>
        </div>

        <v-divider vertical />

        <div class="pl-content">
          <div v-if="!selectedList" class="pa-6 text-medium-emphasis">
            Select a price list on the left, or add one.
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
              <v-btn
                icon="mdi-pencil"
                size="small"
                variant="text"
                @click="openEditList"
              />
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
                v-model="addPrice"
                density="compact"
                hide-details
                label="Price"
                prefix="$"
                style="max-width: 140px;"
                type="number"
                variant="outlined"
                @keyup.enter="addItem"
              />
              <v-btn
                color="primary"
                :disabled="!addProductId || addPrice === ''"
                :loading="addingItem"
                text="Add"
                variant="flat"
                @click="addItem"
              />
            </div>

            <v-alert
              v-if="itemError"
              class="mx-4 mb-2"
              density="compact"
              type="error"
              variant="tonal"
            >
              {{ itemError }}
            </v-alert>

            <v-data-table
              class="pl-table"
              density="comfortable"
              :headers="itemHeaders"
              :items="items"
              :items-per-page="-1"
              :loading="itemsLoading"
            >
              <template #item.price="{ item }">${{ formatMoney(item.price) }}</template>
              <template #item.actions="{ item }">
                <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEditItem(item)" />
                <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="removeItem(item)" />
              </template>
              <template #bottom />
            </v-data-table>
          </template>
        </div>
      </div>
    </v-card>

    <!-- create / edit list -->
    <v-dialog v-model="listDialog" max-width="520">
      <v-card>
        <v-card-title>{{ editingList ? 'Edit Price List' : 'Add Price List' }}</v-card-title>
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

    <!-- edit item price -->
    <v-dialog v-model="editItemDialog" max-width="420">
      <v-card>
        <v-card-title>{{ editItemTarget?.product_code }} — {{ editItemTarget?.product_name }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="editItemPrice"
            autofocus
            density="compact"
            hide-details
            label="Price"
            prefix="$"
            type="number"
            variant="outlined"
            @keyup.enter="saveEditItem"
          />
        </v-card-text>
        <v-card-actions>
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
        <v-card-title>Delete price list</v-card-title>
        <v-card-text>
          Delete <strong>{{ selectedList?.name }}</strong> and all its custom prices? This can't be undone.
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
      kind="price list"
      :load-fn="() => fetchPriceListBranchIds(selectedList.id)"
      :save-fn="ids => setPriceListBranches(selectedList.id, ids)"
      :title="`Branches using “${selectedList?.name}”`"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import BranchAssignDialog from '@/components/BranchAssignDialog.vue'
  import { authFetch } from '@/composables/useApi'
  import { usePriceLists } from '@/composables/usePriceLists'
  import { API_BASE } from '@/config'
  import { formatMoney } from '@/utils/money'

  const {
    priceLists,
    fetchPriceLists,
    createPriceList,
    updatePriceList,
    deletePriceList,
    fetchPriceListItems,
    savePriceListItem,
    deletePriceListItem,
    fetchPriceListBranchIds,
    setPriceListBranches,
  } = usePriceLists()

  const loadError = ref('')
  const selection = ref([])
  const selectedList = computed(() => priceLists.value.find(p => p.id === selection.value[0]) ?? null)

  const itemHeaders = [
    { title: 'Code', key: 'product_code', align: 'start', width: 140 },
    { title: 'Product', key: 'product_name', align: 'start' },
    { title: 'Price', key: 'price', align: 'end', width: 140 },
    { title: '', key: 'actions', align: 'end', sortable: false, width: 100 },
  ]

  const items = ref([])
  const itemsLoading = ref(false)
  const itemError = ref('')

  async function loadItems () {
    if (!selectedList.value) {
      items.value = []
      return
    }
    itemsLoading.value = true
    itemError.value = ''
    try {
      items.value = await fetchPriceListItems(selectedList.value.id)
    } catch (error) {
      itemError.value = error.message
    } finally {
      itemsLoading.value = false
    }
  }

  watch(selection, loadItems)

  onMounted(async () => {
    try {
      await fetchPriceLists(true)
      if (selection.value.length === 0 && priceLists.value.length > 0) {
        selection.value = [priceLists.value[0].id]
      }
    } catch (error) {
      loadError.value = error.message
    }
  })

  // ---- product search --------------------------------------------------
  const productOptions = ref([])
  const productSearchLoading = ref(false)
  const addProductId = ref(null)
  const addPrice = ref('')
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
      if (query && query.trim()) {
        url.searchParams.set('search', query.trim())
      }
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

  function toCents (dollars) {
    return Math.round(Number.parseFloat(dollars) * 100)
  }

  async function addItem () {
    if (!addProductId.value || addPrice.value === '') return
    const cents = toCents(addPrice.value)
    if (!Number.isFinite(cents) || cents < 0) {
      itemError.value = 'Enter a valid price.'
      return
    }
    addingItem.value = true
    itemError.value = ''
    try {
      await savePriceListItem({ priceListId: selectedList.value.id, productId: addProductId.value, price: cents })
      addProductId.value = null
      addPrice.value = ''
      await loadItems()
    } catch (error) {
      itemError.value = error.message
    } finally {
      addingItem.value = false
    }
  }

  // ---- edit / delete item -------------------------------------------------
  const editItemDialog = ref(false)
  const editItemTarget = ref(null)
  const editItemPrice = ref('')
  const savingItem = ref(false)

  function openEditItem (item) {
    editItemTarget.value = item
    editItemPrice.value = (item.price / 100).toString()
    editItemDialog.value = true
  }

  async function saveEditItem () {
    const cents = toCents(editItemPrice.value)
    if (!Number.isFinite(cents) || cents < 0) return
    savingItem.value = true
    try {
      await savePriceListItem({ priceListId: selectedList.value.id, productId: editItemTarget.value.product_id, price: cents })
      editItemDialog.value = false
      await loadItems()
    } catch (error) {
      itemError.value = error.message
    } finally {
      savingItem.value = false
    }
  }

  async function removeItem (item) {
    try {
      await deletePriceListItem(selectedList.value.id, item.product_id)
      await loadItems()
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
        await updatePriceList({ id: selectedList.value.id, ...payload })
      } else {
        const created = await createPriceList(payload)
        if (created?.id) selection.value = [created.id]
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
      await deletePriceList(selectedList.value.id)
      deleteListDialog.value = false
      selection.value = priceLists.value.length > 0 ? [priceLists.value[0].id] : []
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
.page-root, .pl-card {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}
.pl-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.pl-sidebar {
  width: 240px;
  flex: 0 0 240px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}
.pl-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
</style>
