<template>
  <v-dialog v-model="editorDialog" fullscreen scrollable>
    <v-card>
      <v-card-title class="d-flex align-center ga-3">
        <span>{{ editingId ? `Stock Count #${editingId}` : 'New Stock Count' }}</span>
        <v-chip v-if="editorStatus" :color="statusColor(editorStatus)" size="small">{{ editorStatus }}</v-chip>
        <v-spacer />
        <v-btn icon="mdi-close" variant="text" @click="closeEditor" />
      </v-card-title>
      <v-divider />
      <v-card-text class="pa-4">
        <v-row>
          <v-col cols="6" md="3">
            <v-select
              v-model="inventoryId"
              density="compact"
              :disabled="!isDraft"
              hide-details
              :item-title="i => i.name"
              :item-value="i => i.id"
              :items="inventories"
              label="Inventory"
              @update:model-value="seedFromInventory"
            />
          </v-col>
          <v-col cols="6" md="6">
            <v-text-field v-model="note" density="compact" :disabled="!isDraft" hide-details label="Note" />
          </v-col>
          <v-col v-if="isDraft" cols="12" md="3">
            <v-autocomplete
              v-model="addVariantId"
              density="compact"
              hide-details
              hide-no-data
              :item-title="v => v.label"
              :item-value="v => v.id"
              :items="skuOptions"
              :loading="skuLoading"
              no-filter
              placeholder="Add SKU by search / barcode"
              @update:model-value="addLine"
              @update:search="onSkuSearch"
            />
          </v-col>
        </v-row>

        <v-alert v-if="editorError" class="my-2" type="error" variant="tonal">{{ editorError }}</v-alert>

        <div v-if="seedProgress" class="d-flex align-center ga-2 my-2">
          <v-progress-linear color="primary" :model-value="seedProgress.pct" />
          <span class="text-caption text-medium-emphasis text-no-wrap">
            {{ seedProgress.loaded }} / {{ seedProgress.total }} SKUs
          </span>
        </div>

        <!-- Virtualized: only the rows on screen ever mount, so a count of
             thousands of SKUs stays smooth instead of one input per row. -->
        <v-data-table-virtual
          class="lines-table mt-2"
          fixed-header
          :headers="lineHeaders"
          height="calc(100vh - 340px)"
          :items="lines"
          item-value="variantId"
        >
          <template #item.system_quantity="{ item }">
            <div class="text-end">{{ item.systemQuantity }}</div>
          </template>
          <template #item.counted_quantity="{ item }">
            <v-text-field
              v-if="isDraft"
              v-model.number="item.countedQuantity"
              density="compact"
              hide-details
              reverse
              style="max-width: 120px; margin-left: auto;"
              type="number"
            />
            <div v-else class="text-end">{{ item.countedQuantity }}</div>
          </template>
          <template #item.difference="{ item }">
            <div class="text-end" :class="diffClass(item)">{{ item.countedQuantity - item.systemQuantity }}</div>
          </template>
          <template #item.actions="{ item }">
            <v-icon-btn
              v-if="isDraft"
              icon="mdi-delete"
              size="small"
              variant="text"
              @click="lines.splice(lines.findIndex(l => l.variantId === item.variantId), 1)"
            />
          </template>
          <template #no-data>
            <div class="text-medium-emphasis pa-4">No lines yet.</div>
          </template>
        </v-data-table-virtual>
      </v-card-text>
      <v-divider />
      <v-card-actions class="pa-4">
        <v-btn
          v-if="isDraft"
          color="error"
          :disabled="!editingId"
          :loading="cancelling"
          text="Cancel Count"
          variant="text"
          @click="cancelDraft"
        />
        <v-spacer />
        <template v-if="isDraft">
          <v-btn :loading="saving" text="Save Draft" @click="save" />
          <v-btn color="success" :disabled="!editingId" :loading="posting" text="Post" @click="post" />
        </template>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="counts-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-clipboard-list-outline" />
        <span>Stock Counts</span>
        <v-spacer />
        <v-select
          v-model="filterInventoryId"
          class="filter-select"
          clearable
          density="compact"
          hide-details
          :item-title="i => i.name"
          :item-value="i => i.id"
          :items="inventories"
          label="Inventory"
          variant="outlined"
        />
        <v-btn color="primary" prepend-icon="mdi-plus" text="New Count" variant="flat" @click="openCreate" />
      </v-card-title>
      <v-divider />

      <ServerSideTable
        ref="tableRef"
        :api-u-r-l="apiURL"
        density="comfortable"
        flush
        :headers="headers"
        hover
        :query-params="{ inventory_id: filterInventoryId }"
        root-key="stock_counts"
        :show-search-icon="false"
        @row-click="openExisting"
      >
        <template #item.status="{ item }">
          <v-chip :color="statusColor(item.status)" size="small">{{ item.status }}</v-chip>
        </template>
        <template #item.created_at="{ item }">
          {{ new Date(item.created_at).toLocaleString() }}
        </template>
      </ServerSideTable>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import { useRoute } from 'vue-router'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { useStockCounts } from '@/composables/useStockCounts'
  import { useVariants, variantLabel } from '@/composables/useVariants'
  import { API_BASE } from '@/config'

  const route = useRoute()
  const { inventories, fetchInventories } = useInventories()
  const {
    createStockCount, updateStockCount, fetchStockCount,
    postStockCount, cancelStockCount, fetchCountSheet,
  } = useStockCounts()
  const { variants: skuOptions, loading: skuLoading, searchVariants } = useVariants()
  fetchInventories()
  searchVariants('')

  let skuSearchTimer = null
  function onSkuSearch (q) {
    clearTimeout(skuSearchTimer)
    skuSearchTimer = setTimeout(() => searchVariants(q || ''), 250)
  }

  const apiURL = `${API_BASE}/stock_counts`
  const filterInventoryId = ref(null)
  const headers = [
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Inventory', key: 'inventory_name', align: 'start' },
    { title: 'Note', key: 'note', align: 'start' },
    { title: 'Status', key: 'status', align: 'start' },
    { title: 'Created At', key: 'created_at', align: 'end' },
  ]

  function statusColor (s) {
    return { draft: 'grey', posted: 'success', cancelled: 'error' }[s] || 'grey'
  }

  const tableRef = ref(null)

  const editorDialog = ref(false)
  const editingId = ref(null)
  const editorStatus = ref('draft')
  const inventoryId = ref(null)
  const note = ref('')
  const lines = ref([])
  const addVariantId = ref(null)
  const editorError = ref('')
  const saving = ref(false)
  const posting = ref(false)
  const cancelling = ref(false)

  const isDraft = computed(() => editorStatus.value === 'draft')

  const lineHeaders = computed(() => [
    { title: 'SKU', key: 'label' },
    { title: 'System Qty', key: 'system_quantity', align: 'end' },
    { title: 'Counted Qty', key: 'counted_quantity', align: 'end' },
    { title: 'Difference', key: 'difference', align: 'end' },
    ...(isDraft.value ? [{ title: '', key: 'actions', align: 'end', width: 40 }] : []),
  ])

  function diffClass (line) {
    const diff = line.countedQuantity - line.systemQuantity
    if (diff > 0) return 'text-success'
    if (diff < 0) return 'text-error'
    return ''
  }

  const seedProgress = ref(null)
  let seedToken = 0

  function openCreate () {
    seedToken++ // cancel any in-flight seed from a previous open
    seedProgress.value = null
    editingId.value = null
    editorStatus.value = 'draft'
    inventoryId.value = null
    note.value = ''
    lines.value = []
    editorError.value = ''
    editorDialog.value = true
  }

  // Arriving from Inventory's "Start Count" button pre-selects that
  // inventory and seeds the sheet from its live stock.
  const preselectInventoryId = route.query.inventory_id ? Number(route.query.inventory_id) : null
  if (preselectInventoryId) {
    openCreate()
    inventoryId.value = preselectInventoryId
    seedFromInventory(preselectInventoryId)
  }

  async function openExisting (item) {
    seedToken++ // cancel any in-flight seed from a previous open
    seedProgress.value = null
    editorError.value = ''
    try {
      const data = await fetchStockCount(item.id)
      const count = data?.stock_count ?? item
      editingId.value = count.id
      editorStatus.value = count.status
      inventoryId.value = count.inventory_id
      note.value = count.note || ''
      lines.value = (data?.items ?? []).map(it => ({
        variantId: it.variant_id,
        label: [it.product_code, it.product_name].filter(Boolean).join(' — '),
        systemQuantity: it.system_quantity,
        countedQuantity: it.counted_quantity,
      }))
      editorDialog.value = true
    } catch (error) {
      editorError.value = error.message
    }
  }

  function closeEditor () {
    editorDialog.value = false
  }

  function toLine (s) {
    return {
      variantId: s.variant_id,
      label: [s.product_code, s.product_name].filter(Boolean).join(' — '),
      systemQuantity: s.quantity,
      countedQuantity: s.quantity,
    }
  }

  // Seeding a new draft: pull every SKU currently stocked in the chosen
  // inventory as the starting count sheet, counted = system to start.
  // Loaded and rendered in batches (see fetchCountSheet) so a large
  // inventory streams in instead of blocking on one giant fetch.
  async function seedFromInventory (id) {
    if (!id || editingId.value) return
    editorError.value = ''
    lines.value = []
    const token = ++seedToken
    seedProgress.value = { loaded: 0, total: 0, pct: 0 }
    try {
      await fetchCountSheet(id, {
        onPage: (rows, { loaded, total }) => {
          if (token !== seedToken) return // a newer selection superseded this one
          lines.value.push(...rows.map(toLine))
          seedProgress.value = { loaded, total, pct: total ? (loaded / total) * 100 : 100 }
        },
      })
    } catch (error) {
      if (token === seedToken) editorError.value = error.message
    } finally {
      if (token === seedToken) seedProgress.value = null
    }
  }

  function addLine (variantId) {
    if (!variantId) return
    if (lines.value.some(l => l.variantId === variantId)) {
      addVariantId.value = null
      return
    }
    const variant = skuOptions.value.find(v => v.id === variantId)
    lines.value.push({
      variantId,
      label: variant ? variantLabel(variant) : `SKU #${variantId}`,
      systemQuantity: 0,
      countedQuantity: 0,
    })
    addVariantId.value = null
  }

  function toItems () {
    return lines.value.map(l => ({ variantId: l.variantId, countedQuantity: Number(l.countedQuantity) || 0 }))
  }

  async function save () {
    if (!inventoryId.value) {
      editorError.value = 'Select an inventory.'
      return
    }
    saving.value = true
    editorError.value = ''
    try {
      const saved = editingId.value
        ? await updateStockCount({ id: editingId.value, inventoryId: inventoryId.value, note: note.value, items: toItems() })
        : await createStockCount({ inventoryId: inventoryId.value, note: note.value, items: toItems() })
      editingId.value = saved?.id ?? editingId.value
      editorStatus.value = saved?.status ?? editorStatus.value
      tableRef.value?.reload()
    } catch (error) {
      editorError.value = error.message
    } finally {
      saving.value = false
    }
  }

  async function post () {
    if (!editingId.value) return
    posting.value = true
    editorError.value = ''
    try {
      await save()
      const res = await postStockCount(editingId.value)
      editorStatus.value = res?.stock_count?.status ?? 'posted'
      tableRef.value?.reload()
    } catch (error) {
      editorError.value = error.message
    } finally {
      posting.value = false
    }
  }

  async function cancelDraft () {
    if (!editingId.value) return
    cancelling.value = true
    editorError.value = ''
    try {
      const cancelled = await cancelStockCount(editingId.value)
      editorStatus.value = cancelled?.status ?? 'cancelled'
      tableRef.value?.reload()
    } catch (error) {
      editorError.value = error.message
    } finally {
      cancelling.value = false
    }
  }
</script>

<style scoped>
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.counts-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.filter-select {
  flex: 0 0 200px;
}
</style>
