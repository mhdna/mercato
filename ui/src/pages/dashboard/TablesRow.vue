<template>
  <div style="flex: 1 1 0; min-height: 200px">
    <!-- <v-text-field
      v-model="search"
      clearable
      density="compact"
      hide-details
      label="Search invoices..."
      prepend-inner-icon="mdi-magnify"
      single-line
      class="my-2 mx-12"
    /> -->
    <div class="d-flex">
      <v-card
        class="flex-1 d-flex flex-column w-100"
        flat
        style="min-width: 0; min-height: 0"
      >
        <ServerSideTable
          :api-u-r-l="`${API_BASE}/sales_invoices`"
          :external-search="search"
          fill-height
          :headers="invoiceHeaders"
          root-key="sales_invoices"
          title="Sales Invoices"
        />
      </v-card>
      <v-card
        class="flex-1 d-flex flex-column w-100"
        flat
        style="min-width: 0; min-height: 0"
      >
        <ServerSideTable
          :api-u-r-l="`${API_BASE}/purchases`"
          :external-search="search"
          fill-height
          :headers="purchaseHeaders"
          root-key="purchases"
          title="Purchase Invoices"
        />
      </v-card>
    </div>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { API_BASE } from '@/config'

  const search = ref('')

  const invoiceHeaders = ref([
    { title: 'Code', key: 'invoice_code', align: 'start' },
    { title: 'Client ID', key: 'client_id', align: 'start' },
    { title: 'Subtotal', key: 'subtotal', align: 'end' },
    { title: 'Discount', key: 'discount', align: 'end' },
    { title: 'Grand Total', key: 'grand_total', align: 'end' },
  ])

  const purchaseHeaders = ref([
    { title: 'Supplier ID', key: 'supplier_id', align: 'start' },
    { title: 'Purchased At', key: 'purchased_at', align: 'end' },
  ])
</script>
