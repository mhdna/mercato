<template>
  <div class="d-flex flex-column loyalty-board">
    <v-card class="d-flex flex-column pyramid-card" flat>
      <v-card-title class="d-flex flex-wrap align-center justify-space-between ga-2">
        <div class="d-flex align-center">
          <v-icon class="me-2" color="amber-darken-2" icon="mdi-trophy" />
          Top Spenders
        </div>
        <div class="d-flex align-center ga-2">
          <v-select
            v-model="sortBy"
            density="compact"
            hide-details
            :items="sortOptions"
            label="Sort by"
            style="width: 180px"
            variant="outlined"
          />
          <v-select
            v-model="pyramidCount"
            density="compact"
            hide-details
            :items="[20, 50, 100]"
            label="Show top"
            style="width: 110px"
            variant="outlined"
          />
          <v-icon-btn icon="mdi-cog-outline" title="Edit tiers" variant="text" @click="tiersDialog = true" />
        </div>
      </v-card-title>

      <div class="d-flex flex-wrap ga-2 px-4 pb-3">
        <v-chip
          v-for="t in sortedTiers()"
          :key="t.id"
          class="text-white"
          :color="t.color"
          size="small"
          variant="flat"
        >
          {{ t.label }} · ${{ t.minAmount }}+
        </v-chip>
      </div>

      <v-card-text class="pyramid-scroll flex-grow-1">
        <div v-if="pyramidLoading" class="d-flex justify-center pa-8">
          <v-progress-circular color="primary" indeterminate />
        </div>
        <div v-else-if="pyramidClients.length === 0" class="text-center text-medium-emphasis pa-8">
          No spending data yet.
        </div>
        <div v-else class="pyramid">
          <div v-for="(row, ri) in pyramidRows" :key="ri" class="pyramid-row">
            <div v-for="client in row" :key="client.id" class="pyramid-item">
              <v-avatar class="pyramid-avatar mb-1" :color="tierFor(client.total_spent)?.color" :size="avatarSize(ri)">
                <span class="text-white font-weight-bold" :style="{ fontSize: `${avatarSize(ri) / 2.6}px` }">
                  {{ initials(client.name) }}
                </span>
              </v-avatar>
              <div class="pyramid-name text-truncate">{{ client.name }}</div>
              <div class="pyramid-amount text-medium-emphasis">{{ metricLabel(client) }}</div>
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <v-card class="d-flex flex-column rest-card mt-4" flat>
      <v-card-title class="d-flex flex-wrap align-center justify-space-between ga-2">
        All Clients
        <v-text-field
          v-model="search"
          clearable
          density="compact"
          hide-details
          label="Search name or phone"
          prepend-inner-icon="mdi-magnify"
          style="max-width: 280px"
          variant="outlined"
        />
      </v-card-title>
      <div class="rest-scroll flex-grow-1">
        <v-data-table-server
          v-model:items-per-page="restItemsPerPage"
          density="compact"
          :headers="restHeaders"
          item-value="id"
          :items="restItems"
          :items-length="restTotal"
          :items-per-page-options="[10, 25, 50, 100]"
          :loading="restLoading"
          @update:options="loadRest"
        >
          <template #item.name="{ item }">
            <router-link :to="`/clients/${item.id}`">{{ item.name }}</router-link>
          </template>
          <template #item.total_spent="{ item }">${{ formatMoney(item.total_spent) }}</template>
          <template #item.invoice_count="{ item }">{{ item.invoice_count }}</template>
          <template #item.item_count="{ item }">{{ item.item_count }}</template>
          <template #item.actions="{ item }">
            <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="$emit('edit', item)" />
            <v-icon-btn
              color="error"
              icon="mdi-delete"
              size="small"
              variant="text"
              @click="$emit('delete', item)"
            />
          </template>
        </v-data-table-server>
      </div>
    </v-card>

    <v-dialog v-model="tiersDialog" max-width="600">
      <v-card>
        <v-card-title class="d-flex align-center justify-space-between">
          Spend Tiers
          <v-icon-btn icon="mdi-close" variant="text" @click="tiersDialog = false" />
        </v-card-title>
        <v-card-text>
          <p class="text-caption text-medium-emphasis mb-4">
            A client's tier is the highest one whose minimum they clear. Shared across the Retail and Wholesale tabs.
          </p>
          <div v-for="t in tiers" :key="t.id" class="d-flex align-center ga-3 mb-3">
            <input v-model="t.color" class="tier-color-input" type="color">
            <v-text-field
              v-model="t.label"
              density="compact"
              hide-details
              label="Name"
              style="max-width: 160px"
            />
            <v-text-field
              v-model.number="t.minAmount"
              density="compact"
              hide-details
              label="Min spend"
              prefix="$"
              style="max-width: 130px"
              type="number"
            />
            <v-spacer />
            <v-icon-btn color="error" icon="mdi-delete" variant="text" @click="removeTier(t.id)" />
          </div>
          <v-btn prepend-icon="mdi-plus" variant="tonal" @click="addTier">Add Tier</v-btn>
        </v-card-text>
        <v-card-actions>
          <v-btn text="Reset to Defaults" variant="text" @click="resetTiers" />
          <v-spacer />
          <v-btn color="primary" text="Done" @click="tiersDialog = false" />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref, watch } from 'vue'
  import { useClientTiers } from '@/composables/useClientTiers'
  import { dedupedFetch } from '@/composables/useRequestDedup'
  import { API_BASE } from '@/config'
  import { formatMoney } from '@/utils/money'

  const props = defineProps({
    branchId: { type: [Number, String, null], default: null },
    clientType: { type: String, required: true },
  })
  defineEmits(['edit', 'delete'])

  const { tiers, sortedTiers, tierFor, addTier, removeTier, resetTiers } = useClientTiers()
  const tiersDialog = ref(false)

  function initials (name) {
    return (name ?? '')
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map(w => w[0]?.toUpperCase())
      .join('')
  }

  function avatarSize (rowIndex) {
    return Math.max(40, 88 - rowIndex * 8)
  }

  // Pyramid shape: row 1 has 1 client, row 2 has 2, etc, so the board
  // visually widens toward the bottom with the top spender alone at the peak.
  const pyramidRows = computed(() => {
    const rows = []
    let i = 0
    let rowSize = 1
    while (i < pyramidClients.value.length) {
      rows.push(pyramidClients.value.slice(i, i + rowSize))
      i += rowSize
      rowSize++
    }
    return rows
  })

  const sortOptions = [
    { title: 'Sort: Most Spent', value: 'spending' },
    { title: 'Sort: Most Invoices', value: 'invoices' },
    { title: 'Sort: Most Items Bought', value: 'items' },
  ]
  const sortBy = ref('spending')

  function metricLabel (client) {
    if (sortBy.value === 'invoices') return `${client.invoice_count} invoice${client.invoice_count === 1 ? '' : 's'}`
    if (sortBy.value === 'items') return `${client.item_count} item${client.item_count === 1 ? '' : 's'}`
    return `$${formatMoney(client.total_spent)}`
  }

  const pyramidCount = ref(20)
  const pyramidClients = ref([])
  const pyramidLoading = ref(true)

  function baseParams () {
    const params = new URLSearchParams()
    params.set('client_type', props.clientType)
    params.set('sort_by', sortBy.value)
    if (props.branchId) params.set('branch_id', props.branchId)
    return params
  }

  async function fetchPyramid () {
    pyramidLoading.value = true
    try {
      const params = baseParams()
      params.set('page_size', pyramidCount.value)
      params.set('page_id', 0)
      const data = await dedupedFetch(`${API_BASE}/clients/by_spending?${params}`)
      pyramidClients.value = data?.clients ?? []
    } catch {
      pyramidClients.value = []
    } finally {
      pyramidLoading.value = false
    }
  }

  const search = ref('')
  const restItemsPerPage = ref(10)
  const restItems = ref([])
  const restTotal = ref(0)
  const restLoading = ref(true)
  // Sorting is driven by the "Sort by" dropdown (server-side, matches the
  // pyramid above), not by clicking column headers, so header sort arrows
  // are turned off to avoid implying a sort that wouldn't actually happen.
  const restHeaders = [
    { title: 'Name', key: 'name', align: 'start', sortable: false },
    { title: 'Phone', key: 'phone', align: 'start', sortable: false },
    { title: 'Total Spent', key: 'total_spent', align: 'end', sortable: false },
    { title: 'Invoices', key: 'invoice_count', align: 'end', sortable: false },
    { title: 'Items Bought', key: 'item_count', align: 'end', sortable: false },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ]

  const lastRestOptions = ref({ page: 1, itemsPerPage: 10 })
  async function loadRest (options) {
    lastRestOptions.value = options
    restLoading.value = true
    try {
      const searchActive = !!search.value
      const { page, itemsPerPage } = options
      // The rest-of-clients table skips the clients already shown in the
      // pyramid above it -- unless the admin is searching, in which case
      // they're looking for one specific client regardless of rank.
      const offset = searchActive
        ? (page - 1) * itemsPerPage
        : (page - 1) * itemsPerPage + pyramidCount.value

      const params = baseParams()
      if (searchActive) params.set('search', search.value)
      params.set('page_size', itemsPerPage)
      params.set('page_id', offset)

      const data = await dedupedFetch(`${API_BASE}/clients/by_spending?${params}`)
      restItems.value = data?.clients ?? []
      const total = data?.total ?? 0
      restTotal.value = searchActive ? total : Math.max(total - pyramidCount.value, 0)
    } catch {
      restItems.value = []
      restTotal.value = 0
    } finally {
      restLoading.value = false
    }
  }

  function reload () {
    fetchPyramid()
    loadRest({ ...lastRestOptions.value, page: 1 })
  }

  defineExpose({ reload })

  onMounted(fetchPyramid)
  watch(() => [props.clientType, props.branchId, pyramidCount.value, sortBy.value], reload)

  let searchTimeout = null
  watch(search, () => {
    clearTimeout(searchTimeout)
    searchTimeout = setTimeout(() => {
      loadRest({ ...lastRestOptions.value, page: 1 })
    }, 300)
  })
</script>

<style scoped>
.loyalty-board {
  height: calc(100vh - 260px);
  min-height: 560px;
}

.pyramid-card {
  flex: 1 1 50%;
  min-height: 0;
  overflow: hidden;
}

.rest-card {
  flex: 1 1 50%;
  min-height: 0;
  overflow: hidden;
}

.pyramid-scroll {
  overflow-y: auto;
  min-height: 0;
}

.rest-scroll {
  overflow-y: auto;
  min-height: 0;
}

.pyramid-row {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-bottom: 20px;
}

.pyramid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100px;
}

.pyramid-avatar {
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25);
}

.pyramid-name {
  max-width: 96px;
  font-size: 12px;
  font-weight: 500;
  text-align: center;
}

.pyramid-amount {
  font-size: 11px;
}

.tier-color-input {
  inline-size: 36px;
  block-size: 36px;
  border: none;
  background: none;
  padding: 0;
  cursor: pointer;
}
</style>
