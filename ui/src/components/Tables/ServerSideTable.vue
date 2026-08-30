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
      v-model:items-per-page="itemsPerPage"
      v-model:page="page"
      :density="density"
      :headers="headers"
      :hover="hover"
      item-value="id"
      :items="serverItems"
      :items-length="totalItems"
      :items-per-page-options="itemsPerPageOptions"
      :loading="loading"
      @click:row="onRowClick"
      @update:options="loadItemsTracked"
    >
      <template v-for="(_, slotName) of $slots" :key="slotName" #[slotName]="scope">
        <slot :name="slotName" v-bind="scope" />
      </template>
    </v-data-table-server>
  </v-card>
</template>
<script setup>
  import { ref, watch } from 'vue'
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
    title: {
      type: String,
      default: '',
    },
    totalKey: {
      type: String,
      default: 'total',
    },
  })

  const emit = defineEmits(['row-click'])

  // The table needs all valid choices, not only the initial page size.

  const itemsPerPageOption_ = [...new Set([props.defaultItemsPerPage, 14, 25, 50, 100])]
    .find(n => n <= props.maxPageSize)
  const itemsPerPageOption = props.defaultItemsPerPage <= props.maxPageSize
    ? props.defaultItemsPerPage
    : itemsPerPageOption_
  const itemsPerPage = ref(itemsPerPageOption ?? 14)
  const page = ref(1)
  const serverItems = ref([])
  const loading = ref(true)
  const totalItems = ref(-1)
  const error = ref('')

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
    applyQueryParams(url)

    const data = await dedupedFetch(url.toString())

    let items = Array.isArray(data) ? data : (data[props.rootKey] ?? [])
    const total = Array.isArray(data) ? -1 : (data[props.totalKey] ?? -1)

    if (sortBy.length > 0) {
      const sortKey = sortBy[0].key
      const sortOrder = sortBy[0].order
      items = items.toSorted((a, b) => {
        const aValue = a[sortKey]
        const bValue = b[sortKey]
        return sortOrder === 'desc'
          ? (bValue > aValue ? 1 : -1)
          : (aValue > bValue ? 1 : -1)
      })
    }

    return { items, total }
  }

  function loadItems ({ page, itemsPerPage, sortBy }) {
    loading.value = true
    error.value = ''
    fetchPage({ page, itemsPerPage, sortBy })
      .then(({ items, total }) => {
        serverItems.value = items
        totalItems.value = total
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

  defineExpose({ reload })
</script>
