<template>
  <ServerSideTable
    :api-u-r-l="`${API_BASE}/invoices`"
    :external-search="search"
    fill-height
    :headers="invoiceHeaders"
    root-key="invoices"
    title="Revenue Analysis"
  />
  <ServerSideTable
    :api-u-r-l="`${API_BASE}/invoices`"
    :external-search="search"
    fill-height
    :headers="invoiceHeaders"
    root-key="invoices"
    title="Slow Movers"
  />
  <ServerSideTable
    :api-u-r-l="`${API_BASE}/invoices`"
    :external-search="search"
    fill-height
    :headers="invoiceHeaders"
    root-key="invoices"
    title="Fast Movers"
  />
  <ServerSideTable
    :api-u-r-l="`${API_BASE}/invoices`"
    :external-search="search"
    fill-height
    :headers="invoiceHeaders"
    root-key="invoices"
    title="Reordering Suggestions"
  />
</template>

<script setup>
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { API_BASE } from '@/config'

  const search = ref('')

  const invoiceHeaders = ref([
    // { title: "ID", key: "id", align: "start" },
    { title: 'Code', key: 'invoice_code', align: 'start' },
    // { title: "Type", key: "type", align: "start" },
    { title: 'Branch', key: 'branch', align: 'start' },
    { title: 'Currency', key: 'currency_code', align: 'start' },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Discount', key: 'discount', align: 'end' },
    { title: 'Net', key: 'netamount', align: 'end' },
  // { title: "Client", key: "client_name", align: "start" },
  ])

  const purchaseHeaders = ref([
    // { title: "ID", key: "id", align: "start" },
    { title: 'Code', key: 'invoice_code', align: 'start' },
    // { title: "Category", key: "type", align: "start" },
    { title: 'Branch', key: 'branch', align: 'start' },
    { title: 'Currency', key: 'currency_code', align: 'start' },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Discount', key: 'discount', align: 'end' },
    { title: 'Net', key: 'netamount', align: 'end' },
  // { title: "Client", key: "client_name", align: "start" },
  ])
</script>
