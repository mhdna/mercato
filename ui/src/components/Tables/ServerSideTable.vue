<template>
  <v-card :class="flush ? 'px-0 py-0' : 'px-4 py-4'" flat>
    <v-card-title
      v-if="props.title || props.showSearchIcon"
      class="card-title d-flex justify-space-between"
    >
      {{ props.title }}
      <v-icon v-if="props.showSearchIcon" icon="mdi-magnify" size="26" />
    </v-card-title>
    <v-alert
      v-if="error"
      class="mb-2"
      closable
      type="error"
      variant="tonal"
      @click:close="error = ''"
    >
      {{ error === "Server is down, we'll be back soon." ? error : `Failed to load data: ${error}` }}
      <template #append>
        <v-btn size="small" text="Retry" variant="text" @click="reload" />
      </template>
    </v-alert>
    <v-data-table-server
      v-model:expanded="expandedRows"
      v-model:items-per-page="itemsPerPage"
      v-model:page="page"
      v-model:selected="selected"
      :density="density"
      :expand-on-click="expandable"
      :headers="resolvedHeaders"
      :hover="hover"
      :item-value="itemValue"
      :items="serverItems"
      :items-length="totalItems"
      :items-per-page-options="itemsPerPageOptions"
      :loading="loading"
      :show-expand="expandable"
      :show-select="selectable"
      @click:row="onRowClick"
      @update:options="loadItemsTracked"
    >
      <template v-for="(_, slotName) of $slots" :key="slotName" #[slotName]="scope">
        <slot :name="slotName" v-bind="scope" />
      </template>
    </v-data-table-server>
  </v-card>
</template>
<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { dedupedFetch } from '@/composables/useRequestDedup'

  const props = defineProps({
    apiURL: {
      type: String,
      required: true,
    },
    externalSearch: {
      type: String,
      default: '',
    },
    // Extra query params merged into every request URL. Empty strings,
    // null/undefined and empty arrays are skipped; arrays are sent as
    // repeated params (?k=1&k=2). Changing this refetches from page 1.
    queryParams: {
      type: Object,
      default: () => ({}),
    },
    defaultItemsPerPage: {
      type: Number,
      default: 14,
    },
    density: {
      type: String,
      default: 'compact',
    },
    flush: {
      type: Boolean,
      default: false,
    },
    showSearchIcon: {
      type: Boolean,
      default: true,
    },
    fillHeight: {
      type: Boolean,
      default: false,
    },
    hover: {
      type: Boolean,
      default: false,
    },
    headers: {
      type: Array,
      required: true,
    },
    // Column keys the server can sort on (its sort_by whitelist). A header
    // whose key is listed sorts server-side via ?sort_by=&sort_order=;
    // every other column has its sort arrow suppressed, because this table
    // only ever holds one page and a client-side sort of that page would
    // silently reorder 14 rows instead of the whole result set.
    sortKeys: {
      type: Array,
      default: () => [],
    },
    maxPageSize: {
      type: Number,
      default: 100,
    },
    searchDebounce: {
      type: Number,
      default: 100,
    },
    rootKey: {
      type: String,
      required: true,
    },
    // Enables the checkbox column + header select-all. The parent reads the
    // selection via the exposed `selected` ref / `update:selected` event.
    selectable: {
      type: Boolean,
      default: false,
    },
    // Enables the expand column + expand-on-click. Pair with an
    // #expanded-row slot; bind :expanded / @update:expanded for the ids.
    expandable: {
      type: Boolean,
      default: false,
    },
    expanded: {
      type: Array,
      default: () => [],
    },
    itemValue: {
      type: String,
      default: 'id',
    },
    title: {
      type: String,
      default: '',
    },
    totalKey: {
      type: String,
      default: 'total',
    },
  })

  const emit = defineEmits(['row-click', 'update:selected', 'update:expanded', 'loaded'])

  const expandedRows = ref([...props.expanded])
  watch(() => props.expanded, value => {
    if (value !== expandedRows.value) expandedRows.value = [...value]
  })
  watch(expandedRows, value => emit('update:expanded', value), { deep: true })

  const sortableKeys = new Set(props.sortKeys)
  const resolvedHeaders = computed(() =>
    props.headers.map(h => ({
      ...h,
      sortable: h.sortable ?? sortableKeys.has(h.key),
    })),
  )

  // The table needs all valid choices, not only the initial page size.
  // Every option (and the initial value) must stay within maxPageSize, or
  // the server rejects the request with a PageSize "max" validation error.
  const itemsPerPageOptions = [
    ...new Set([props.defaultItemsPerPage, 14, 25, 50, 100, props.maxPageSize]),
  ]
    .filter(n => n > 0 && n <= props.maxPageSize)
    .toSorted((a, b) => a - b)
  const itemsPerPage = ref(
    props.defaultItemsPerPage <= props.maxPageSize
      ? props.defaultItemsPerPage
      : (itemsPerPageOptions.at(-1) ?? props.maxPageSize),
  )
  const page = ref(1)
  const serverItems = ref([])
  const loading = ref(true)
  const totalItems = ref(-1)
  const error = ref('')
  const selected = ref([])

  watch(selected, value => emit('update:selected', value), { deep: true })

  function clearSelection () {
    selected.value = []
  }

  function applyQueryParams (url) {
    for (const [key, value] of Object.entries(props.queryParams || {})) {
      if (value === null || value === undefined || value === '') continue
      if (Array.isArray(value)) {
        for (const v of value) {
          if (v !== null && v !== undefined && v !== '') url.searchParams.append(key, v)
        }
        continue
      }
      url.searchParams.set(key, value)
    }
  }

  async function fetchPage ({ page, itemsPerPage, sortBy }) {
    const offset = (page - 1) * itemsPerPage
    const url = new URL(props.apiURL)
    url.searchParams.set('page_size', itemsPerPage)
    url.searchParams.set('page_id', offset)
    if (props.externalSearch.trim()) {
      url.searchParams.set('search', props.externalSearch.trim())
    }
    // Server-side sort only, and only for whitelisted columns (see sortKeys).
    if (sortBy.length > 0 && sortableKeys.has(sortBy[0].key)) {
      url.searchParams.set('sort_by', sortBy[0].key)
      url.searchParams.set('sort_order', sortBy[0].order || 'asc')
    }
    applyQueryParams(url)

    const data = await dedupedFetch(url.toString())

    const items = Array.isArray(data) ? data : (data[props.rootKey] ?? [])
    const total = Array.isArray(data) ? -1 : (data[props.totalKey] ?? -1)

    return { items, total }
  }

  function loadItems ({ page, itemsPerPage, sortBy }) {
    loading.value = true
    error.value = ''
    fetchPage({ page, itemsPerPage, sortBy })
      .then(({ items, total }) => {
        serverItems.value = items
        totalItems.value = total
        // Collapse rows that fell off the page and tell the parent what
        // loaded (so it can drop any per-row state it was caching).
        if (expandedRows.value.length > 0) {
          const ids = new Set(items.map(i => i[props.itemValue]))
          expandedRows.value = expandedRows.value.filter(id => ids.has(id))
        }
        emit('loaded', items)
      })
      .catch(error_ => {
        error.value = error_.message
        serverItems.value = []
        totalItems.value = 0
      })
      .finally(() => {
        loading.value = false
      })
  }

  const lastOptions = ref(null)
  function loadItemsTracked (options) {
    lastOptions.value = options
    loadItems(options)
  }

  function reload () {
    if (lastOptions.value) loadItems(lastOptions.value)
  }

  // Search / filter changes must land the user back on page 1 — otherwise
  // the narrowed result set is paged from wherever they happened to be.
  function reloadFromFirstPage () {
    if (page.value !== 1) {
      page.value = 1 // triggers @update:options -> loadItemsTracked
      return
    }
    if (lastOptions.value) {
      lastOptions.value = { ...lastOptions.value, page: 1 }
      loadItems(lastOptions.value)
    }
  }

  let debounceTimer = null
  function debouncedReload () {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(reloadFromFirstPage, props.searchDebounce)
  }

  function onRowClick (_event, { item }) {
    emit('row-click', item)
  }

  watch(() => props.externalSearch, debouncedReload)
  watch(() => props.queryParams, debouncedReload, { deep: true })

  defineExpose({ reload, reloadFromFirstPage, selected, clearSelection, items: serverItems })
</script>
