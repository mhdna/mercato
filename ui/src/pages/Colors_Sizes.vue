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

    <v-card class="cs-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-palette" />
        <span>Colors &amp; Sizes</span>
        <v-spacer />
        <v-text-field
          v-if="tab === 'colors'"
          v-model="colorSearch"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search colors"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-text-field
          v-else
          v-model="sizeSearch"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search sizes"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn
          v-if="tab === 'colors'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Color"
          variant="flat"
          @click="openAddColor"
        />
        <v-btn
          v-else
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Size"
          variant="flat"
          @click="openAddSize"
        />
      </v-card-title>
      <v-divider />

      <div class="cs-layout">
        <v-tabs
          v-model="tab"
          class="cs-tabs"
          color="primary"
          direction="vertical"
        >
          <v-tab prepend-icon="mdi-palette" text="Colors" value="colors" />
          <v-tab prepend-icon="mdi-tshirt-crew" text="Sizes" value="sizes" />
        </v-tabs>

        <v-divider vertical />

        <div class="cs-content">
          <!-- ---------------------------------------------------------- COLORS -->
          <div v-show="tab === 'colors'">
            <div class="pa-4">
              <BulkDeleteBar
                :count="colorSelected.length"
                :error="colorSelError"
                :loading="colorDeleting"
                @clear="resetColorSel"
                @confirm="bulkDeleteColors"
              />
              <ServerSideTable
                ref="colorTableRef"
                :api-u-r-l="`${API_BASE}/colors`"
                class="color-table"
                density="comfortable"
                :external-search="colorSearch ?? ''"
                flush
                :headers="colorHeaders"
                item-value="id"
                root-key="colors"
                selectable
                :show-search-icon="false"
                :sort-keys="['id', 'name', 'hex_value', 'created_at']"
                @row-click="openEditColor"
                @update:selected="v => (colorSelected = v)"
              >
                <template #item.hex_value="{ item }">
                  <div class="d-flex align-center ga-2">
                    <div class="swatch" :style="{ backgroundColor: item.hex_value }" />
                    <span class="text-caption">{{ item.hex_value }}</span>
                  </div>
                </template>
                <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
              </ServerSideTable>
            </div>
          </div>

          <!-- ----------------------------------------------------------- SIZES -->
          <div v-show="tab === 'sizes'">
            <div class="pa-4">
              <BulkDeleteBar
                :count="sizeSelected.length"
                :error="sizeSelError"
                :loading="sizeDeleting"
                @clear="resetSizeSel"
                @confirm="bulkDeleteSizes"
              />
              <ServerSideTable
                ref="sizeTableRef"
                :api-u-r-l="`${API_BASE}/sizes`"
                class="size-table"
                density="comfortable"
                :external-search="sizeSearch ?? ''"
                flush
                :headers="sizeHeaders"
                item-value="id"
                root-key="sizes"
                selectable
                :show-search-icon="false"
                :sort-keys="['id', 'name', 'type', 'order', 'created_at']"
                @row-click="openEditSize"
                @update:selected="v => (sizeSelected = v)"
              >
                <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
              </ServerSideTable>
            </div>
          </div>
        </div>
      </div>
    </v-card>

    <!-- add color -->
    <v-dialog v-model="addColorDialog" max-width="360">
      <v-card rounded="lg">
        <v-card-title class="d-flex align-center ga-2 px-4 py-3 text-h6">
          <v-icon icon="mdi-palette" size="22" />
          <span>Add Color</span>
          <v-spacer />
          <v-btn icon="mdi-close" size="small" variant="text" @click="addColorDialog = false" />
        </v-card-title>
        <v-divider />
        <v-card-text class="px-4 pt-3">
          <v-color-picker
            v-model="colorForm.hex"
            class="color-dialog-picker mb-3"
            elevation="0"
            hide-inputs
            mode="hex"
            width="100%"
          />
          <v-text-field
            v-model="colorForm.name"
            density="compact"
            hide-details="auto"
            label="Color name"
            variant="outlined"
          />
          <v-text-field
            v-model="colorForm.hex"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Hex"
            variant="outlined"
          >
            <template #append-inner>
              <div class="swatch" :style="{ backgroundColor: isValidHex(colorForm.hex) ? colorForm.hex : 'transparent' }" />
            </template>
          </v-text-field>
          <v-alert
            v-if="colorError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ colorError }}</v-alert>
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="addColorDialog = false" />
          <v-btn
            color="primary"
            :loading="colorSaving"
            text="Add"
            variant="flat"
            @click="addColor"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- add size -->
    <v-dialog v-model="addSizeDialog" max-width="420">
      <v-card>
        <v-card-title class="text-h6">Add Size</v-card-title>
        <v-card-text>
          <v-combobox
            v-model="sizeForm.type"
            class="mb-3"
            density="compact"
            hide-details="auto"
            :items="sizeTypes"
            label="Type"
            variant="outlined"
          />
          <v-text-field
            v-model="sizeForm.name"
            density="compact"
            hide-details="auto"
            label="Name"
            placeholder="XL / 42"
            variant="outlined"
          />
          <v-text-field
            v-model="sizeForm.order"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Order"
            type="number"
            variant="outlined"
          />
          <v-alert
            v-if="sizeError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ sizeError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="addSizeDialog = false" />
          <v-btn
            color="primary"
            :loading="sizeSaving"
            text="Add"
            variant="flat"
            @click="addSize"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- edit color -->
    <v-dialog v-model="editColorDialog" max-width="360">
      <v-card rounded="lg">
        <v-card-title class="d-flex align-center ga-2 px-4 py-3 text-h6">
          <v-icon icon="mdi-palette" size="22" />
          <span>Edit Color</span>
          <v-spacer />
          <v-btn icon="mdi-close" size="small" variant="text" @click="editColorDialog = false" />
        </v-card-title>
        <v-divider />
        <v-card-text class="px-4 pt-3">
          <v-color-picker
            v-model="editColorForm.hex"
            class="color-dialog-picker mb-3"
            elevation="0"
            hide-inputs
            mode="hex"
            width="100%"
          />
          <v-text-field
            v-model="editColorForm.name"
            density="compact"
            hide-details="auto"
            label="Name"
            variant="outlined"
          />
          <v-text-field
            v-model="editColorForm.hex"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Hex"
            variant="outlined"
          >
            <template #append-inner>
              <div class="swatch" :style="{ backgroundColor: isValidHex(editColorForm.hex) ? editColorForm.hex : 'transparent' }" />
            </template>
          </v-text-field>
          <v-alert
            v-if="editError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ editError }}</v-alert>
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-btn
            color="error"
            prepend-icon="mdi-delete"
            text="Delete"
            variant="text"
            @click="deleteEditedColor"
          />
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="editColorDialog = false" />
          <v-btn
            color="primary"
            :loading="editSaving"
            text="Save"
            variant="flat"
            @click="saveColor"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- edit size -->
    <v-dialog v-model="editSizeDialog" max-width="420">
      <v-card>
        <v-card-title class="text-h6">Edit Size</v-card-title>
        <v-card-text>
          <v-combobox
            v-model="editSizeForm.type"
            class="mb-3"
            density="compact"
            hide-details="auto"
            :items="sizeTypes"
            label="Type"
            variant="outlined"
          />
          <v-text-field
            v-model="editSizeForm.name"
            density="compact"
            hide-details="auto"
            label="Name"
            variant="outlined"
          />
          <v-text-field
            v-model="editSizeForm.order"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Order"
            type="number"
            variant="outlined"
          />
          <v-alert
            v-if="editError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ editError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-btn
            color="error"
            prepend-icon="mdi-delete"
            text="Delete"
            variant="text"
            @click="deleteEditedSize"
          />
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="editSizeDialog = false" />
          <v-btn
            color="primary"
            :loading="editSaving"
            text="Save"
            variant="flat"
            @click="saveSize"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- delete -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="text-h6">Delete {{ deleteKind }}</v-card-title>
        <v-card-text>
          Delete <strong>{{ deleteLabel }}</strong>? This can't be undone.
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
            @click="doDelete"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useColorsAndSizes } from '@/composables/useColorsAndSizes'
  import { API_BASE } from '@/config'

  const {
    sizes,
    fetchSizes,
    createColor,
    updateColor,
    deleteColor,
    createSize,
    updateSize,
    deleteSize,
  } = useColorsAndSizes()

  const {
    selected: colorSelected,
    deleting: colorDeleting,
    error: colorSelError,
    reset: resetColorSel,
    run: runColorBulk,
  } = useBulkDelete()
  const {
    selected: sizeSelected,
    deleting: sizeDeleting,
    error: sizeSelError,
    reset: resetSizeSel,
    run: runSizeBulk,
  } = useBulkDelete()

  const colorTableRef = ref(null)
  const sizeTableRef = ref(null)
  function reloadColors () {
    colorTableRef.value?.reload()
  }
  function reloadSizes () {
    sizeTableRef.value?.reload()
    // Keep the size-type picker in sync with newly added/removed types.
    fetchSizes(true)
  }

  async function bulkDeleteColors () {
    try {
      await runColorBulk('/colors/bulk_delete', { ids: colorSelected.value })
      resetColorSel()
      reloadColors()
    } catch { /* error shown in the bar */ }
  }

  async function bulkDeleteSizes () {
    try {
      await runSizeBulk('/sizes/bulk_delete', { ids: sizeSelected.value })
      resetSizeSel()
      reloadSizes()
    } catch { /* error shown in the bar */ }
  }

  const tab = ref('colors')
  const loadError = ref('')
  const colorSearch = ref('')
  const sizeSearch = ref('')

  const colorHeaders = [
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Hex', key: 'hex_value', align: 'start' },
    { title: 'ID', key: 'id', align: 'start', width: 80 },
    { title: 'Created At', key: 'created_at', align: 'start', width: 200 },
  ]

  const sizeHeaders = [
    { title: 'Type', key: 'type', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Order', key: 'order', align: 'start', width: 90 },
    { title: 'ID', key: 'id', align: 'start', width: 80 },
    { title: 'Created At', key: 'created_at', align: 'start', width: 200 },
  ]

  // Distinct size types for the add/edit dialog picker — derived from the
  // full size list (fetchSizes), not the visible table page.
  const sizeTypes = computed(() => [...new Set(sizes.value.map(s => s.type))].toSorted())

  onMounted(() => {
    fetchSizes().catch(error => { loadError.value = error.message })
  })

  function isValidHex (value) {
    return /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(value ?? '')
  }

  // v-color-picker can hand back an 8-digit #RRGGBBAA once the alpha slider is
  // touched; the API only accepts #RGB / #RRGGBB, so drop the alpha byte.
  function normalizeHex (value) {
    return typeof value === 'string' && value.length === 9 ? value.slice(0, 7) : value
  }

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
  }

  // ---- add color --------------------------------------------------------
  const colorForm = reactive({ name: '', hex: '#1976D2' })
  const addColorDialog = ref(false)
  const colorSaving = ref(false)
  const colorError = ref('')

  function openAddColor () {
    colorError.value = ''
    Object.assign(colorForm, { name: '', hex: '#1976D2' })
    addColorDialog.value = true
  }

  async function addColor () {
    if (!colorForm.name.trim()) {
      colorError.value = 'Name is required.'
      return
    }
    if (!isValidHex(normalizeHex(colorForm.hex))) {
      colorError.value = 'Pick a valid hex color.'
      return
    }
    colorSaving.value = true
    colorError.value = ''
    try {
      await createColor({ name: colorForm.name.trim(), hexValue: normalizeHex(colorForm.hex) })
      addColorDialog.value = false
      reloadColors()
    } catch (error) {
      colorError.value = error.message
    } finally {
      colorSaving.value = false
    }
  }

  // ---- add size -------------------------------------------------------
  const sizeForm = reactive({ name: '', type: '', order: '' })
  const addSizeDialog = ref(false)
  const sizeSaving = ref(false)
  const sizeError = ref('')

  function openAddSize () {
    sizeError.value = ''
    Object.assign(sizeForm, { name: '', type: '', order: '' })
    addSizeDialog.value = true
  }

  async function addSize () {
    if (!sizeForm.type?.trim() || !sizeForm.name.trim() || sizeForm.order === '') {
      sizeError.value = 'Type, name and order are all required.'
      return
    }
    sizeSaving.value = true
    sizeError.value = ''
    try {
      await createSize({ name: sizeForm.name.trim(), type: sizeForm.type.trim(), order: sizeForm.order })
      addSizeDialog.value = false
      reloadSizes()
    } catch (error) {
      sizeError.value = error.message
    } finally {
      sizeSaving.value = false
    }
  }

  // ---- edit -----------------------------------------------------------
  const editColorDialog = ref(false)
  const editSizeDialog = ref(false)
  const editColorForm = reactive({ id: null, name: '', hex: '' })
  const editSizeForm = reactive({ id: null, name: '', type: '', order: '' })
  const editSaving = ref(false)
  const editError = ref('')

  function openEditColor (item) {
    editError.value = ''
    Object.assign(editColorForm, { id: item.id, name: item.name, hex: item.hex_value })
    editColorDialog.value = true
  }

  function deleteEditedColor () {
    if (editColorForm.id == null) return
    editColorDialog.value = false
    confirmDelete('color', { id: editColorForm.id, name: editColorForm.name })
  }

  function openEditSize (item) {
    editError.value = ''
    Object.assign(editSizeForm, { id: item.id, name: item.name, type: item.type, order: item.order })
    editSizeDialog.value = true
  }

  function deleteEditedSize () {
    if (editSizeForm.id == null) return
    editSizeDialog.value = false
    confirmDelete('size', { id: editSizeForm.id, name: editSizeForm.name, type: editSizeForm.type })
  }

  async function saveColor () {
    if (!editColorForm.name.trim() || !isValidHex(normalizeHex(editColorForm.hex))) {
      editError.value = 'Name and a valid hex are required.'
      return
    }
    editSaving.value = true
    editError.value = ''
    try {
      await updateColor({ id: editColorForm.id, name: editColorForm.name.trim(), hexValue: normalizeHex(editColorForm.hex) })
      editColorDialog.value = false
      reloadColors()
    } catch (error) {
      editError.value = error.message
    } finally {
      editSaving.value = false
    }
  }

  async function saveSize () {
    if (!editSizeForm.type?.trim() || !editSizeForm.name.trim() || editSizeForm.order === '') {
      editError.value = 'Type, name and order are all required.'
      return
    }
    editSaving.value = true
    editError.value = ''
    try {
      await updateSize({
        id: editSizeForm.id,
        name: editSizeForm.name.trim(),
        type: editSizeForm.type.trim(),
        order: editSizeForm.order,
      })
      editSizeDialog.value = false
      reloadSizes()
    } catch (error) {
      editError.value = error.message
    } finally {
      editSaving.value = false
    }
  }

  // ---- delete -------------------------------------------------------
  const deleteDialog = ref(false)
  const deleting = ref(false)
  const deleteError = ref('')
  const deleteKind = ref('')
  const deleteTarget = ref(null)

  const deleteLabel = computed(() => {
    if (!deleteTarget.value) return ''
    return deleteKind.value === 'color'
      ? deleteTarget.value.name
      : `${deleteTarget.value.type}: ${deleteTarget.value.name}`
  })

  function confirmDelete (kind, item) {
    deleteKind.value = kind
    deleteTarget.value = item
    deleteError.value = ''
    deleteDialog.value = true
  }

  async function doDelete () {
    deleting.value = true
    deleteError.value = ''
    try {
      await (deleteKind.value === 'color'
        ? deleteColor(deleteTarget.value.id)
        : deleteSize(deleteTarget.value.id))
      deleteDialog.value = false
      if (deleteKind.value === 'color') reloadColors()
      else reloadSizes()
    } catch (error) {
      deleteError.value = error.message
    } finally {
      deleting.value = false
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
.cs-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.cs-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.cs-tabs {
  flex: 0 0 160px;
}
.cs-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.swatch {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 1px solid rgba(128, 128, 128, 0.6);
}
.table-search {
  flex: 0 1 320px;
}
.color-dialog-picker {
  max-width: 300px;
  margin-inline: auto;
}
:deep(.color-table tbody tr),
:deep(.size-table tbody tr) {
  cursor: pointer;
}
</style>
