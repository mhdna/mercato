<script setup>
  import { computed, ref } from 'vue'
  import WarehouseTreemap from '@/components/Stats/WarehouseTreemap.vue'
  import { warehouseOptions, warehouseTotals } from '@/data/warehouseCosts'

  const options = ref([...warehouseOptions])
  const selected = ref('all')

  const totals = computed(() => warehouseTotals(selected.value))
  const totalCost = computed(() => '$ ' + totals.value.cost.toLocaleString('en-US'))
  const totalWeight = computed(() => totals.value.weight.toLocaleString('en-US') + ' KG')

  const dialog = ref(false)
  const newName = ref('')

  function openAdd () {
    newName.value = ''
    dialog.value = true
  }

  function addWarehouse () {
    const name = newName.value.trim()
    if (!name) return
    const value = name.toLowerCase().replace(/\s+/g, '-')
    options.value.push({ title: name, value })
    selected.value = value
    dialog.value = false
  }
</script>

<template>
  <div class="warehouses-page">
    <div class="filters d-flex align-center ga-3 pa-3">
      <v-select
        v-model="selected"
        density="compact"
        hide-details
        :items="options"
        label="Warehouse"
        max-width="260"
        variant="outlined"
      />

      <v-spacer />
      <div class="text-center">
        <div class="text-h6 font-weight-bold">{{ totalCost }}</div>
        <div class="text-caption text-medium-emphasis">
          Total cost of goods &middot; {{ totalWeight }}
        </div>
      </div>
      <v-spacer />

      <v-btn
        color="primary"
        prepend-icon="mdi-plus"
        text="Add Warehouse"
        @click="openAdd"
      />
    </div>

    <div class="chart-area">
      <WarehouseTreemap :warehouse="selected" />
    </div>
  </div>

  <v-dialog v-model="dialog" max-width="420">
    <v-card>
      <v-card-title>Add Warehouse</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="newName"
          autofocus
          density="compact"
          label="Warehouse name"
          @keyup.enter="addWarehouse"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" @click="dialog = false" />
        <v-btn color="primary" text="Add" @click="addWarehouse" />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.warehouses-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}

.filters {
  flex: 0 0 auto;
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.chart-area {
  flex: 1 1 auto;
  min-height: 0;
}
</style>
