<template>
  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Stock Movements</h2>
  </div>

  <v-row class="mb-1" dense>
    <v-col cols="12" sm="4">
      <v-select
        v-model="filters.inventoryId"
        clearable
        density="compact"
        hide-details
        :item-title="i => i.name"
        :item-value="i => i.id"
        :items="inventories"
        label="Inventory"
        @update:model-value="reload"
      />
    </v-col>
    <v-col cols="12" sm="4">
      <v-select
        v-model="filters.reason"
        clearable
        density="compact"
        hide-details
        :items="reasons"
        label="Reason"
        @update:model-value="reload"
      />
    </v-col>
  </v-row>

  <v-alert v-if="error" class="mb-2" type="error" variant="tonal">{{ error }}</v-alert>

  <v-data-table-server
    density="compact"
    :headers="headers"
    :items="rows"
    :items-length="total"
    :items-per-page="pageSize"
    :loading="loading"
    :page="page"
    @update:options="onOptions"
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
      {{ item.unit_cost?.Valid ? item.unit_cost.Int64 : (item.unit_cost ?? '—') }}
    </template>
    <template #item.created_at="{ item }">
      {{ new Date(item.created_at).toLocaleString() }}
    </template>
  </v-data-table-server>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { useInventories } from '@/composables/useInventories'
  import { STOCK_MOVEMENT_REASONS, useStockMovements } from '@/composables/useStockMovements'

  const { inventories, fetchInventories } = useInventories()
  const { fetchMovements } = useStockMovements()

  const reasons = STOCK_MOVEMENT_REASONS
  const headers = [
    { title: 'When', key: 'created_at' },
    { title: 'Inventory', key: 'inventory_name' },
    { title: 'SKU', key: 'sku', sortable: false },
    { title: 'Barcode', key: 'barcode' },
    { title: 'Reason', key: 'reason' },
    { title: 'Qty', key: 'quantity', align: 'end' },
    { title: 'Unit Cost', key: 'unit_cost', align: 'end' },
    { title: 'Note', key: 'note' },
  ]

  const rows = ref([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref('')
  const page = ref(1)
  const pageSize = ref(50)
  const filters = ref({ inventoryId: null, reason: null })

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

  async function load () {
    loading.value = true
    error.value = ''
    try {
      const { movements, total: t } = await fetchMovements({
        inventoryId: filters.value.inventoryId || undefined,
        reason: filters.value.reason || undefined,
        pageSize: pageSize.value,
        pageId: page.value - 1,
      })
      rows.value = movements
      total.value = t
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  }

  function reload () {
    page.value = 1
    load()
  }

  function onOptions (opts) {
    page.value = opts.page
    pageSize.value = opts.itemsPerPage
    load()
  }

  onMounted(() => {
    fetchInventories()
    load()
  })
</script>
