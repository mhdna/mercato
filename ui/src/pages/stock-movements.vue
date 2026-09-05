<template>
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
      </v-card-title>
      <v-divider />

      <div class="movements-content">
        <div class="pa-4">
          <ServerSideTable
            :api-u-r-l="apiURL"
            density="comfortable"
            flush
            :headers="headers"
            :query-params="{ inventory_id: filters.inventoryId, reason: filters.reason }"
            root-key="movements"
            :show-search-icon="false"
          >
            <template #item.reason="{ item }">
              <v-chip :color="reasonColor(item.reason)" size="small">{{ item.reason }}</v-chip>
            </template>
            <template #item.quantity="{ item }">
              <span :class="item.quantity < 0 ? 'text-error' : 'text-success'">{{ item.quantity }}</span>
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
  import { onMounted, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useInventories } from '@/composables/useInventories'
  import { STOCK_MOVEMENT_REASONS } from '@/composables/useStockMovements'
  import { API_BASE } from '@/config'

  const { inventories, fetchInventories } = useInventories()

  const apiURL = `${API_BASE}/stock_movements`
  const reasons = STOCK_MOVEMENT_REASONS
  const filters = ref({ inventoryId: null, reason: null })

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
    { title: 'Unit Cost', key: 'unit_cost', align: 'end', sortable: false },
    { title: 'Note', key: 'note', sortable: false },
  ]

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
</style>
