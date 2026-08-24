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
          :api-u-r-l="`${API_BASE}/invoices`"
          :external-search="search"
          fill-height
          :headers="invoiceHeaders"
          root-key="invoices"
          title="Sales Invoices"
        />
      </v-card>
      <v-card
        flat
        class="flex-1 d-flex flex-column w-100"
        style="min-width: 0; min-height: 0"
      >
        <ServerSideTable
          :api-u-r-l="`${API_BASE}/invoices`"
          :external-search="search"
          fill-height
          :headers="purchaseHeaders"
          root-key="invoices"
          title="Purchase Invoices"
        />
      </v-card>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import ServerSideTable from "@/components/Tables/ServerSideTable.vue";
import { API_BASE } from "@/config";

const search = ref("");

const invoiceHeaders = ref([
  // { title: "ID", key: "id", align: "start" },
  { title: "Code", key: "invoice_code", align: "start" },
  // { title: "Type", key: "type", align: "start" },
  { title: "Branch", key: "branch", align: "start" },
  { title: "Currency", key: "currency_code", align: "start" },
  { title: "Amount", key: "amount", align: "end" },
  { title: "Discount", key: "discount", align: "end" },
  { title: "Net", key: "netamount", align: "end" },
  // { title: "Client", key: "client_name", align: "start" },
]);

const purchaseHeaders = ref([
  // { title: "ID", key: "id", align: "start" },
  { title: "Code", key: "invoice_code", align: "start" },
  // { title: "Category", key: "type", align: "start" },
  { title: "Branch", key: "branch", align: "start" },
  { title: "Currency", key: "currency_code", align: "start" },
  { title: "Amount", key: "amount", align: "end" },
  { title: "Discount", key: "discount", align: "end" },
  { title: "Net", key: "netamount", align: "end" },
  // { title: "Client", key: "client_name", align: "start" },
]);
</script>
