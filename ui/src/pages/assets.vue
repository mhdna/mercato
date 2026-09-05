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

    <v-card class="assets-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-hammer-wrench" />
        <span>Assets</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search assets"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn
          color="primary"
          :disabled="activeCategories.length === 0"
          prepend-icon="mdi-plus"
          text="Add Asset"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <div class="assets-layout">
        <CategorySidebar
          v-model="selectedCategory"
          :categories="assetCategories"
          :group-by-scope="false"
          @add="openCreateCategory"
          @delete="removeCategory"
          @edit="openEditCategory"
        />

        <v-divider vertical />

        <div class="assets-content">
          <div class="pa-4">
            <BulkDeleteBar
              :count="selectedRows.length"
              :error="selError"
              :loading="bulkDeleting"
              @clear="resetSel"
              @confirm="bulkDeleteRows"
            />
            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="`${API_BASE}/assets/`"
              class="asset-table"
              density="comfortable"
              :external-search="search ?? ''"
              flush
              :headers="headers"
              item-value="id"
              :query-params="{ category_id: selectedCategory }"
              root-key="assets"
              selectable
              :show-search-icon="false"
              :sort-keys="['id', 'name', 'code', 'bought_at', 'created_at']"
              @row-click="openEdit"
              @update:selected="v => (selectedRows = v)"
            >
              <template #item.category_id="{ item }">
                <CategoryChip :category="categoryFor(item.category_id)" />
              </template>
              <template #item.bought_at="{ item }">{{ formatDate(item.bought_at) }}</template>
              <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
            </ServerSideTable>
          </div>
        </div>
      </div>
    </v-card>

    <!-- add / edit asset -->
    <v-dialog v-model="editDialog" max-width="460">
      <v-card>
        <v-card-title class="text-h6">{{ editing ? 'Edit Asset' : 'Add Asset' }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="form.name"
            autofocus
            class="mb-3"
            density="compact"
            hide-details="auto"
            label="Name"
            variant="outlined"
          />
          <v-text-field
            v-model="form.code"
            class="mb-3"
            density="compact"
            hide-details="auto"
            label="Code"
            variant="outlined"
          />
          <v-select
            v-model="form.categoryId"
            class="mb-3"
            density="compact"
            hide-details="auto"
            item-title="name"
            item-value="id"
            :items="activeCategories"
            label="Category"
            variant="outlined"
          >
            <template #item="{ props: itemProps, item }">
              <v-list-item v-bind="itemProps">
                <template #prepend>
                  <v-avatar :color="categoryColor(item.raw)" rounded="md" size="24" variant="tonal">
                    <v-icon :color="categoryColor(item.raw)" :icon="categoryIcon(item.raw)" size="14" />
                  </v-avatar>
                </template>
              </v-list-item>
            </template>
          </v-select>
          <v-text-field
            v-model="form.boughtAt"
            density="compact"
            hide-details="auto"
            label="Bought At"
            type="date"
            variant="outlined"
          />
          <v-alert
            v-if="formError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ formError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-btn
            v-if="editing"
            color="error"
            prepend-icon="mdi-delete"
            text="Delete"
            variant="text"
            @click="deleteEditedAsset"
          />
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="editDialog = false" />
          <v-btn
            color="primary"
            :loading="saving"
            text="Save"
            variant="flat"
            @click="saveAsset"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- delete asset -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="text-h6">Delete asset</v-card-title>
        <v-card-text>
          Delete <strong>{{ deleteTarget?.name }}</strong>? This can't be undone.
          <v-alert
            v-if="deleteError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ deleteError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="deleteDialog = false" />
          <v-btn
            color="error"
            :loading="deleting"
            text="Delete"
            variant="flat"
            @click="doDelete"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <CategoryDialog
      v-model="categoryDialog"
      :category="editingCategory"
      name-hint="e.g. Vehicles, Equipment, Furniture"
      :on-submit="submitCategory"
      :show-scope="false"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import CategoryChip from '@/components/CategoryChip.vue'
  import CategorySidebar from '@/components/CategorySidebar.vue'
  import CategoryDialog from '@/components/Forms/CategoryDialog.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useAssetCategories } from '@/composables/useAssetCategories'
  import { useAssets } from '@/composables/useAssets'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { API_BASE } from '@/config'
  import { categoryColor, categoryIcon } from '@/data/expenseCategoryIcons'

  const {
    assetCategories,
    fetchAssetCategories,
    createAssetCategory,
    updateAssetCategory,
    deleteAssetCategory,
  } = useAssetCategories()
  const { createAsset, updateAsset, deleteAsset } = useAssets()

  const {
    selected: selectedRows,
    deleting: bulkDeleting,
    error: selError,
    reset: resetSel,
    run: runBulk,
  } = useBulkDelete()

  const tableRef = ref(null)
  function reloadRows () {
    tableRef.value?.reload()
  }

  async function bulkDeleteRows () {
    try {
      await runBulk('/assets/bulk_delete', { ids: selectedRows.value })
      resetSel()
      reloadRows()
    } catch { /* error shown in the bar */ }
  }

  const loadError = ref('')
  const search = ref('')
  const selectedCategory = ref(null)

  const activeCategories = computed(() => assetCategories.value.filter(c => c.is_active))

  function categoryFor (id) {
    return assetCategories.value.find(c => c.id === id) ?? null
  }

  const headers = [
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Category', key: 'category_id', align: 'start' },
    { title: 'Bought At', key: 'bought_at', align: 'start', width: 160 },
    { title: 'Created At', key: 'created_at', align: 'start', width: 200 },
  ]

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleDateString()
  }

  function toDateInput (value) {
    if (!value) return ''
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? '' : d.toISOString().slice(0, 10)
  }

  onMounted(() => {
    fetchAssetCategories(true).catch(error => { loadError.value = error.message })
  })

  watch(selectedCategory, () => resetSel())

  // ---- create / edit asset ----------------------------------------------
  const editDialog = ref(false)
  const editing = ref(null)
  const saving = ref(false)
  const formError = ref('')
  const form = reactive({ name: '', code: '', categoryId: null, boughtAt: '' })

  function openCreate () {
    editing.value = null
    formError.value = ''
    Object.assign(form, {
      name: '',
      code: '',
      categoryId: selectedCategory.value ?? activeCategories.value[0]?.id ?? null,
      boughtAt: toDateInput(new Date()),
    })
    editDialog.value = true
  }

  function openEdit (item) {
    editing.value = item
    formError.value = ''
    Object.assign(form, {
      name: item.name,
      code: item.code,
      categoryId: item.category_id,
      boughtAt: toDateInput(item.bought_at),
    })
    editDialog.value = true
  }

  async function saveAsset () {
    if (!form.name.trim() || !form.code.trim()) {
      formError.value = 'Name and code are required.'
      return
    }
    if (!form.categoryId) {
      formError.value = 'Pick a category.'
      return
    }
    saving.value = true
    formError.value = ''
    const payload = {
      name: form.name.trim(),
      code: form.code.trim(),
      categoryId: form.categoryId,
      boughtAt: form.boughtAt ? new Date(form.boughtAt).toISOString() : '',
    }
    try {
      await (editing.value
        ? updateAsset({ id: editing.value.id, ...payload })
        : createAsset(payload))
      editDialog.value = false
      reloadRows()
    } catch (error) {
      formError.value = error.message
    } finally {
      saving.value = false
    }
  }

  // ---- delete asset ----------------------------------------------------
  const deleteDialog = ref(false)
  const deleteTarget = ref(null)
  const deleting = ref(false)
  const deleteError = ref('')

  function deleteEditedAsset () {
    if (!editing.value) return
    deleteTarget.value = editing.value
    deleteError.value = ''
    editDialog.value = false
    deleteDialog.value = true
  }

  async function doDelete () {
    deleting.value = true
    deleteError.value = ''
    try {
      await deleteAsset(deleteTarget.value.id)
      deleteDialog.value = false
      reloadRows()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
    }
  }

  // ---- categories ----------------------------------------------------
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
      ? updateAssetCategory({ id: editingCategory.value.id, ...payload })
      : createAssetCategory(payload))
    await fetchAssetCategories(true)
    reloadRows()
  }

  async function removeCategory (category) {
    if (!confirm(`Delete category "${category.name}"? Categories still used by an asset can't be deleted.`)) return
    try {
      await deleteAssetCategory(category.id)
      if (selectedCategory.value === category.id) selectedCategory.value = null
      await fetchAssetCategories(true)
    } catch (error) {
      loadError.value = error.message
    }
  }
</script>

<style scoped>
/* Fill the layout's scroll wrapper and keep the scroll inside the table,
   so the sidebar and page header stay put. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.assets-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.assets-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.assets-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.table-search {
  flex: 0 1 320px;
}
:deep(.asset-table tbody tr) {
  cursor: pointer;
}
</style>
