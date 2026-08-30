<template>
  <ServerSideTable
    ref="tableRef"
    :api-u-r-l="apiURL"
    :default-items-per-page="14"
    density="comfortable"
    :external-search="props.search"
    flush
    :headers="headers"
    hover
    :query-params="queryParams"
    root-key="products"
    :show-search-icon="false"
    @row-click="$emit('edit', $event)"
  >
    <template #item.is_active="{ item }">
      <v-icon
        :color="item.is_active ? 'success' : 'disabled'"
        :icon="item.is_active ? 'mdi-check-circle' : 'mdi-close-circle'"
        size="18"
      />
    </template>
    <template #item.created_at="{ item }">
      {{ formatDate(item.created_at) }}
    </template>
  </ServerSideTable>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { API_BASE } from '@/config'

  const props = defineProps({
    search: {
      type: String,
      default: '',
    },
    filters: {
      type: Object,
      default: () => ({}),
    },
  })

  defineEmits(['edit'])

  const apiURL = `${API_BASE}/products`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Active', key: 'is_active', align: 'center' },
    { title: 'Created At', key: 'created_at', align: 'end' },
  ])

  // Only send filters that are actually set; ServerSideTable skips empty
  // values but keeping the object lean avoids needless URL churn.
  const queryParams = computed(() => {
    const f = props.filters || {}
    const out = {}
    if (f.is_active === true || f.is_active === false) out.is_active = String(f.is_active)
    if (f.has_variants === true) out.has_variants = 'true'
    if (f.created_from) out.created_from = f.created_from
    if (f.created_to) out.created_to = f.created_to
    if (Array.isArray(f.attribute_value_ids) && f.attribute_value_ids.length > 0) {
      out.attribute_value_ids = f.attribute_value_ids
    }
    return out
  })

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleDateString()
  }

  const tableRef = ref(null)
  function reload () {
    tableRef.value?.reload()
  }

  defineExpose({ reload })
</script>
