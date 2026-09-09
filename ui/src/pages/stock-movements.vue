<template>
  <v-dialog v-model="detailsDialog" max-width="480">
    <v-card v-if="detailsTarget">
      <v-card-title>Movement #{{ detailsTarget.id }}</v-card-title>
      <v-card-text>
        <v-row density="compact">
          <v-col cols="6"><span class="text-caption">Before</span><div>{{ detailsTarget.quantity_before }}</div></v-col>
          <v-col cols="6"><span class="text-caption">After</span><div>{{ detailsTarget.quantity_after }}</div></v-col>
          <v-col cols="6"><span class="text-caption">Actor</span><div>{{ detailsTarget.created_by_name || '—' }}</div></v-col>
          <v-col cols="6"><span class="text-caption">Reference</span><div>{{ nullableString(detailsTarget.reference_type) || '—' }} #{{ nullableString(detailsTarget.reference_id) || '' }}</div></v-col>
        </v-row>
        <v-divider class="my-2" />
        <div class="text-caption mb-1">Metadata</div>
        <pre class="metadata-block">{{ formattedMetadata }}</pre>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="detailsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="page-root">
    <v-card class="movements-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-swap-vertical" />
        <span>Stock Movements</span>
        <v-spacer />
        <v-select
          v-model="filters.inventoryId"
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
        <v-select
          v-model="filters.reason"
          class="filter-select"
          clearable
          density="compact"
          hide-details
          :items="reasons"
          label="Reason"
          variant="outlined"
        />
        <v-select
          v-model="filters.referenceType"
          class="filter-select"
          clearable
          density="compact"
          hide-details
          :items="referenceTypes"
          label="Document"
          variant="outlined"
        />
      </v-card-title>
      <v-divider />

      <div class="movements-content">
        <div class="pa-4">
          <ServerSideTable
            :api-u-r-l="apiURL"
            density="comfortable"
            flush
            :headers="headers"
            :query-params="{ inventory_id: filters.inventoryId, reason: filters.reason, reference_type: filters.referenceType }"
            root-key="movements"
            :show-search-icon="false"
            @row-click="openDetails"
          >
            <template #item.reason="{ item }">
              <v-chip :color="reasonColor(item.reason)" size="small">{{ item.reason }}</v-chip>
            </template>
            <template #item.quantity="{ item }">
              <span :class="item.quantity < 0 ? 'text-error' : 'text-success'">{{ item.quantity }}</span>
            </template>
            <template #item.change="{ item }">
              {{ item.quantity_before }} → {{ item.quantity_after }}
            </template>
            <template #item.sku="{ item }">
              {{ item.product_code }} — {{ item.product_name }}
              <span v-if="item.color_name || item.size_name" class="text-medium-emphasis">
                ({{ [item.color_name, item.size_name].filter(Boolean).join(' / ') }})
              </span>
            </template>
            <template #item.unit_cost="{ item }">
              {{ displayUnitCost(item.unit_cost) }}
            </template>
            <template #item.created_by_name="{ item }">
              {{ item.created_by_name || '—' }}
            </template>
            <template #item.created_at="{ item }">
              {{ new Date(item.created_at).toLocaleString() }}
            </template>
          </ServerSideTable>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { STOCK_MOVEMENT_REASONS } from '@/composables/useStockMovements'
  import { API_BASE } from '@/config'

  const { inventories, fetchInventories } = useInventories()

  const apiURL = `${API_BASE}/stock_movements`
  const reasons = STOCK_MOVEMENT_REASONS
  const referenceTypes = ['sales_invoice', 'return_invoice', 'purchase', 'transfer', 'stock_count', 'purchase_cancel', 'transfer_cancel']
  const filters = ref({ inventoryId: null, reason: null, referenceType: null })

  function nullableString (value) {
    if (typeof value === 'string') return value
    if (!value || typeof value !== 'object') return ''
    const valid = value.Valid ?? value.valid
    if (valid === false) return ''
    return value.String ?? value.string ?? value.Int64 ?? value.int64 ?? ''
  }

  function displayUnitCost (value) {
    if (value == null) return '—'
    if (typeof value !== 'object') return value

    const valid = value.Valid ?? value.valid
    if (valid === false) return '—'

    return value.Int64 ?? value.int64 ?? '—'
  }

  const headers = [
    { title: 'When', key: 'created_at', sortable: false },
    { title: 'Inventory', key: 'inventory_name', sortable: false },
    { title: 'SKU', key: 'sku', sortable: false },
    { title: 'Barcode', key: 'barcode', sortable: false },
    { title: 'Reason', key: 'reason', sortable: false },
    { title: 'Qty', key: 'quantity', align: 'end', sortable: false },
    { title: 'Before → After', key: 'change', align: 'end', sortable: false },
    { title: 'Unit Cost', key: 'unit_cost', align: 'end', sortable: false },
    { title: 'By', key: 'created_by_name', sortable: false },
    { title: 'Note', key: 'note', sortable: false },
  ]

  const detailsDialog = ref(false)
  const detailsTarget = ref(null)

  function openDetails (item) {
    detailsTarget.value = item
    detailsDialog.value = true
  }

  const formattedMetadata = computed(() => {
    const raw = detailsTarget.value?.metadata
    if (!raw) return '{}'
    try {
      return JSON.stringify(typeof raw === 'string' ? JSON.parse(raw) : raw, null, 2)
    } catch {
      return String(raw)
    }
  })

  function reasonColor (r) {
    return {
      purchase: 'success',
      sale: 'error',
      return: 'warning',
      transfer_out: 'deep-orange',
      transfer_in: 'teal',
      adjustment: 'blue-grey',
      count: 'indigo',
    }[r] || 'grey'
  }

  onMounted(fetchInventories)
</script>

<style scoped>
/* Fill the layout's flex-column scroll wrapper so the scroll lives inside the
   table, not the whole page. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.movements-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.movements-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.filter-select {
  flex: 0 0 200px;
}
.metadata-block {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.8rem;
}
</style>
