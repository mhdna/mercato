<template>
  <div class="d-flex">
    <v-card
      class="pa-4 mx-2 w-100"
      variant="tonal"
    >
      <v-card-title>29</v-card-title>
      <v-card-text>Low Stock</v-card-text>
    </v-card>
    <v-card
      class="pa-4 me-2 w-100"
      variant="tonal"
    >
      <v-card-title>29</v-card-title>
      <v-card-text>Out Of Stock</v-card-text>
    </v-card>
    <v-card
      class="pa-4 me-2 w-100"
      variant="tonal"
    >
      <v-card-title>29</v-card-title>
      <v-card-text>Overstocked</v-card-text>
    </v-card>
    <v-card
      class="pa-4 me-2 w-100"
      variant="tonal"
    >
      <v-card-title>29</v-card-title>
      <v-card-text>Not Sold 60D+</v-card-text>
    </v-card>
  </div>
  <ServerSideTable
    :api-u-r-l="`${API_BASE}/low-stock`"
    :external-search="search"
    fill-height
    :headers="lowStockHeaders"
    root-key="low_stock"
    title="Low Stock"
  />
</template>

<script setup>
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { API_BASE } from '@/config'

  const search = ref('')

  const lowStockHeaders = ref([
    { title: 'Code', key: 'product_code', align: 'start' },
    { title: 'Product', key: 'product_name', align: 'start' },
    { title: 'Branch', key: 'branch', align: 'start' },
    { title: 'Stock Qty', key: 'stock_qty', align: 'end' },
    { title: 'Min Qty', key: 'min_qty', align: 'end' },
    { title: 'Status', key: 'status', align: 'start' },
  ])

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
