<template>
  <div class="d-flex loyalty-board">
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
            :items="[15, 50, 100]"
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
        <v-alert v-else-if="pyramidError" class="ma-4" type="error" variant="tonal">
          {{ pyramidError === "Server is down, we'll be back soon." ? pyramidError : `Failed to load data: ${pyramidError}` }}
        </v-alert>
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
              <div class="pyramid-name text-truncate">{{ firstName(client.name) }}</div>
              <div class="pyramid-amount text-medium-emphasis">{{ metricLabel(client) }}</div>
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <v-divider class="board-divider" vertical />

    <v-card class="d-flex flex-column rest-card" flat>
      <v-card-title class="d-flex flex-wrap align-center justify-space-between ga-2">
        <div class="d-flex align-center ga-2">
          All Clients
          <span class="text-caption text-medium-emphasis">{{ restTotal }}</span>
        </div>
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

      <v-alert v-if="restError" class="ma-4" type="error" variant="tonal">
        {{ restError === "Server is down, we'll be back soon." ? restError : `Failed to load data: ${restError}` }}
        <template #append>
          <v-btn size="small" variant="text" @click="reload">Retry</v-btn>
        </template>
      </v-alert>

      <!-- v-virtual-scroll only mounts the handful of rows on screen, so this
        stays smooth at tens of thousands of clients. Paging is driven by the
        near-bottom check in onRestScroll rather than a pager. -->
      <v-virtual-scroll
        v-else
        ref="restScroller"
        class="rest-scroll flex-grow-1"
        height="100%"
        :item-height="REST_ROW_HEIGHT"
        :items="restItems"
        @scroll.passive="onRestScroll"
      >
        <template #default="{ item }">
          <div
            class="client-row d-flex align-center ga-3 px-4"
            :style="{ height: `${REST_ROW_HEIGHT}px` }"
            @click="openDetail(item)"
          >
            <v-avatar class="flex-shrink-0" :color="avatarColor(item.name)" size="38">
              <span class="text-caption font-weight-bold text-white">{{ initials(item.name) }}</span>
            </v-avatar>

            <div class="client-row-main">
              <div class="d-flex align-center ga-2">
                <span class="text-body-2 font-weight-medium text-truncate">{{ item.name }}</span>
                <span
                  class="tier-dot flex-shrink-0"
                  :style="{ backgroundColor: tierFor(item.total_spent)?.color }"
                  :title="tierFor(item.total_spent)?.label"
                />
              </div>
              <div class="text-caption text-medium-emphasis text-truncate">
                {{ item.phone || '—' }}
              </div>
            </div>

            <div class="client-row-metrics text-end flex-shrink-0">
              <div class="text-body-2 font-weight-medium">${{ formatMoney(item.total_spent) }}</div>
              <div class="text-caption text-medium-emphasis">
                {{ item.invoice_count }} inv · {{ item.item_count }} items
              </div>
            </div>

            <v-icon class="flex-shrink-0 text-disabled" icon="mdi-chevron-right" size="20" />
          </div>
        </template>
      </v-virtual-scroll>

      <div
        v-if="!restError && restLoading && restItems.length === 0"
        class="flex-grow-1 d-flex align-center justify-center"
      >
        <v-progress-circular color="primary" indeterminate size="24" width="2" />
      </div>
      <div
        v-else-if="!restError && !restLoading && restItems.length === 0"
        class="flex-grow-1 d-flex align-center justify-center text-medium-emphasis text-body-2"
      >
        {{ search ? 'No clients match your search.' : 'No clients yet.' }}
      </div>
      <div
        v-else-if="restLoading && restItems.length > 0"
        class="rest-more-loader d-flex align-center justify-center ga-2 py-2 text-caption text-medium-emphasis"
      >
        <v-progress-circular color="primary" indeterminate size="14" width="2" />
        Loading more…
      </div>
    </v-card>

    <v-navigation-drawer
      v-model="detailOpen"
      location="right"
      temporary
      width="380"
    >
      <template v-if="detailClient">
        <div class="pa-4 d-flex align-center ga-3">
          <v-avatar :color="avatarColor(detailClient.name)" size="48">
            <span class="text-subtitle-2 font-weight-bold text-white">{{ initials(detailClient.name) }}</span>
          </v-avatar>
          <div class="flex-grow-1" style="min-width: 0">
            <div class="text-subtitle-1 font-weight-medium text-truncate">{{ detailClient.name }}</div>
            <v-chip
              class="text-white mt-1"
              :color="tierFor(detailClient.total_spent)?.color"
              size="x-small"
              variant="flat"
            >
              {{ tierFor(detailClient.total_spent)?.label }}
            </v-chip>
          </div>
          <v-icon-btn icon="mdi-close" variant="text" @click="detailOpen = false" />
        </div>

        <v-divider />

        <div class="pa-4 d-flex flex-column ga-3">
          <div class="d-flex ga-2">
            <v-sheet class="detail-stat pa-3 rounded" color="surface-light">
              <div class="text-caption text-medium-emphasis">Total spent</div>
              <div class="text-subtitle-1 font-weight-medium">${{ formatMoney(detailClient.total_spent) }}</div>
            </v-sheet>
            <v-sheet class="detail-stat pa-3 rounded" color="surface-light">
              <div class="text-caption text-medium-emphasis">Invoices</div>
              <div class="text-subtitle-1 font-weight-medium">{{ detailClient.invoice_count }}</div>
            </v-sheet>
            <v-sheet class="detail-stat pa-3 rounded" color="surface-light">
              <div class="text-caption text-medium-emphasis">Items</div>
              <div class="text-subtitle-1 font-weight-medium">{{ detailClient.item_count }}</div>
            </v-sheet>
          </div>

          <v-list class="pa-0" density="compact">
            <v-list-item class="px-0" title="Phone">
              <template #append><span class="text-body-2">{{ detailClient.phone || '—' }}</span></template>
            </v-list-item>
            <v-list-item class="px-0" title="Type">
              <template #append><span class="text-body-2 text-capitalize">{{ detailClient.client_type }}</span></template>
            </v-list-item>
            <v-list-item class="px-0" title="Loyalty points">
              <template #append><span class="text-body-2">{{ detailClient.valid_loyalty_points }}</span></template>
            </v-list-item>
            <v-list-item class="px-0" title="Member since">
              <template #append><span class="text-body-2">{{ formatDate(detailClient.created_at) }}</span></template>
            </v-list-item>
          </v-list>

          <v-btn
            append-icon="mdi-open-in-new"
            block
            :to="`/clients/${detailClient.id}`"
            variant="tonal"
          >
            Open full profile
          </v-btn>
          <div class="d-flex ga-2">
            <v-btn
              class="flex-grow-1"
              prepend-icon="mdi-pencil"
              variant="outlined"
              @click="$emit('edit', detailClient); detailOpen = false"
            >
              Edit
            </v-btn>
            <v-btn
              class="flex-grow-1"
              color="error"
              prepend-icon="mdi-delete"
              variant="outlined"
              @click="$emit('delete', detailClient); detailOpen = false"
            >
              Delete
            </v-btn>
          </div>
        </div>
      </template>
    </v-navigation-drawer>

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

  function firstName (name) {
    return (name ?? '').trim().split(/\s+/)[0] ?? ''
  }

  function avatarSize (rowIndex) {
    return Math.max(40, 64 - rowIndex * 4)
  }

  // Pyramid shape: rows widen by two (1, 3, 5, 7, 9) to form the tip, then
  // stay a fixed width so a large "show top 100" stays a tidy centred grid
  // instead of an ever-widening triangle that overflows and wraps raggedly.
  const PYRAMID_MAX_ROW = 9
  const pyramidRows = computed(() => {
    const rows = []
    let i = 0
    let rowSize = 1
    while (i < pyramidClients.value.length) {
      rows.push(pyramidClients.value.slice(i, i + rowSize))
      i += rowSize
      if (rowSize < PYRAMID_MAX_ROW) rowSize += 2
    }
    return rows
  })

  const sortOptions = [
    { title: 'Most Spent', value: 'spending' },
    { title: 'Most Invoices', value: 'invoices' },
    { title: 'Most Items', value: 'items' },
  ]
  const sortBy = ref('spending')

  function metricLabel (client) {
    if (sortBy.value === 'invoices') return `${client.invoice_count} invoice${client.invoice_count === 1 ? '' : 's'}`
    if (sortBy.value === 'items') return `${client.item_count} item${client.item_count === 1 ? '' : 's'}`
    return `$${formatMoney(client.total_spent)}`
  }

  const pyramidCount = ref(50)
  const pyramidClients = ref([])
  const pyramidLoading = ref(true)
  const pyramidError = ref('')

  function baseParams () {
    const params = new URLSearchParams()
    params.set('client_type', props.clientType)
    params.set('sort_by', sortBy.value)
    if (props.branchId) params.set('branch_id', props.branchId)
    return params
  }

  async function fetchPyramid () {
    pyramidLoading.value = true
    pyramidError.value = ''
    try {
      const params = baseParams()
      params.set('page_size', pyramidCount.value)
      params.set('page_id', 0)
      const data = await dedupedFetch(`${API_BASE}/clients/by_spending?${params}`)
      pyramidClients.value = data?.clients ?? []
    } catch (error) {
      pyramidClients.value = []
      pyramidError.value = error.message
    } finally {
      pyramidLoading.value = false
    }
  }

  function avatarColor (name = '') {
    // Deterministic muted colour per name so rows are visually distinct
    // without a loud palette.
    const palette = [
      'indigo-lighten-1', 'teal-lighten-1', 'deep-purple-lighten-1',
      'blue-grey-lighten-1', 'cyan-darken-1', 'green-lighten-1',
      'orange-lighten-1', 'pink-lighten-1',
    ]
    const key = name
    if (!key) return 'grey-lighten-1'
    let hash = 0
    for (let i = 0; i < key.length; i++) hash = Math.trunc(hash * 31 + key.codePointAt(i))
    return palette[Math.abs(hash) % palette.length]
  }

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? '—' : d.toLocaleDateString()
  }

  const search = ref('')
  const REST_ROW_HEIGHT = 64
  const REST_PAGE_SIZE = 40
  const restItems = ref([])
  const restTotal = ref(0)
  const restLoading = ref(false)
  const restError = ref('')
  const restScroller = ref(null)

  // Infinite-scroll bookkeeping. restDone stops us paging past the end;
  // restSeq guards against an older in-flight page overwriting a newer
  // filter/search result.
  let restDone = false
  let restSeq = 0

  async function loadRestPage (reset) {
    if (restLoading.value) return
    if (!reset && restDone) return

    const searchActive = !!search.value
    // Not searching: the list picks up where the "Top Spenders" pyramid
    // leaves off, so the same clients aren't shown twice. Searching: the
    // admin wants one specific client regardless of rank, so show all.
    const base = searchActive ? 0 : pyramidCount.value
    const from = reset ? base : base + restItems.value.length

    restLoading.value = true
    restError.value = ''
    const mySeq = ++restSeq
    try {
      const params = baseParams()
      if (searchActive) params.set('search', search.value)
      params.set('page_size', REST_PAGE_SIZE)
      params.set('page_id', from)

      const data = await dedupedFetch(`${API_BASE}/clients/by_spending?${params}`)
      if (mySeq !== restSeq) return

      const batch = data?.clients ?? []
      const total = data?.total ?? 0
      restTotal.value = searchActive ? total : Math.max(total - pyramidCount.value, 0)
      restItems.value = reset ? batch : [...restItems.value, ...batch]
      restDone = batch.length < REST_PAGE_SIZE || restItems.value.length >= restTotal.value
    } catch (error) {
      if (mySeq !== restSeq) return
      if (reset) restItems.value = []
      restTotal.value = 0
      restError.value = error.message
    } finally {
      if (mySeq === restSeq) restLoading.value = false
    }
  }

  function onRestScroll (event) {
    const el = event.target
    if (el && el.scrollTop + el.clientHeight >= el.scrollHeight - 240) {
      loadRestPage(false)
    }
  }

  function resetRestList () {
    restDone = false
    const el = restScroller.value?.$el
    if (el) el.scrollTop = 0
    loadRestPage(true)
  }

  const detailOpen = ref(false)
  const detailClient = ref(null)
  function openDetail (client) {
    detailClient.value = client
    detailOpen.value = true
  }

  function reload () {
    fetchPyramid()
    resetRestList()
  }

  defineExpose({ reload })

  onMounted(() => {
    fetchPyramid()
    resetRestList()
  })
  watch(() => [props.clientType, props.branchId, pyramidCount.value, sortBy.value], reload)

  let searchTimeout = null
  watch(search, () => {
    clearTimeout(searchTimeout)
    searchTimeout = setTimeout(resetRestList, 300)
  })
</script>

<style scoped>
.loyalty-board {
  height: calc(100vh - 120px);
  min-height: 560px;
}

.pyramid-card {
  flex: 1.7 1 0;
  min-height: 0;
  overflow: hidden;
}

.rest-card {
  flex: 1 1 0;
  min-height: 0;
  overflow: hidden;
}

.pyramid-scroll {
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 0;
}

.pyramid {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-around;
}

.rest-scroll {
  flex: 1 1 0;
  overflow-y: auto;
  min-height: 0;
}

.client-row {
  cursor: pointer;
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  transition: background-color 0.12s ease;
}

.client-row:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.04);
}

.client-row-main {
  flex: 1 1 auto;
  min-width: 0;
}

.client-row-metrics {
  min-width: 96px;
}

.tier-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.rest-more-loader {
  border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.detail-stat {
  flex: 1 1 0;
  text-align: center;
}

.pyramid-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.pyramid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 68px;
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

@media (max-width: 960px) {
  .loyalty-board {
    flex-direction: column;
    height: auto;
  }

  .board-divider {
    display: none;
  }

  .pyramid-card,
  .rest-card {
    min-height: 560px;
  }
}
</style>
