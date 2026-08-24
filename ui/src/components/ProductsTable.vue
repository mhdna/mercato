<template>
  <ServerSideTable ref="tableRef" :api-u-r-l="apiURL" :headers="headers" root-key="products" />
</template>

<script setup>
  import { ref } from 'vue'
  import { API_BASE } from '@/config'

  const apiURL = `${API_BASE}/products`
  const headers = ref([
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Active', key: 'is_active', align: 'center' },
    { title: 'Created At', key: 'created_at', align: 'end' },
  ])

  const tableRef = ref(null)
  function reload () {
    tableRef.value?.reload()
  }

  defineExpose({ reload })
</script>
