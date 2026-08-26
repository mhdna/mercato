<template>
  <v-card class="px-4 py-4" flat>
    <v-card-title class="card-title d-flex justify-space-between">
      {{ props.title }}
      <v-icon icon="mdi-magnify" size="26" />
    </v-card-title>
    <v-alert
      v-if="error"
      class="mb-2"
      closable
      type="error"
      variant="tonal"
      @click:close="error = ''"
    >
      {{ error === 'Server is offline' ? error : `Failed to load data: ${error}` }}
      <template #append>
        <v-btn size="small" text="Retry" variant="text" @click="reload" />
      </template>
    </v-alert>
    <v-data-table-server
      v-model:items-per-page="itemsPerPage"
      density="compact"
      :headers="headers"
      item-value="id"
      :items="serverItems"
      :items-length="totalItems"
      :items-per-page-options="itemsPerPageOptions"
      :loading="loading"
      @update:options="loadItemsTracked"
    >
      <template v-for="(_, slotName) of $slots" :key="slotName" #[slotName]="scope">
        <slot :name="slotName" v-bind="scope" />
      </template>
    </v-data-table-server>
  </v-card>
</template>
<script setup>
  import { ref } from 'vue'
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
    fillHeight: {
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

  // The table needs all valid choices, not only the initial page size.
  // eslint-disable-next-line unicorn/prefer-array-find
  const itemsPerPageOptions = [10, 25, 50, 100].filter(n => n <= props.maxPageSize)
  const itemsPerPageOption = itemsPerPageOptions[0]
  const itemsPerPage = ref(itemsPerPageOption ?? 10)
  const serverItems = ref([])
  const loading = ref(true)
  const totalItems = ref(-1)
  const error = ref('')

  async function fetchPage ({ page, itemsPerPage, sortBy }) {
    const offset = (page - 1) * itemsPerPage
    const url = new URL(props.apiURL)
    url.searchParams.set('page_size', itemsPerPage)
    url.searchParams.set('page_id', offset)

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

  defineExpose({ reload })
</script>
