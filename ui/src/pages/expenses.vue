<template>
  <v-container>
    <v-card>
      <v-card-title class="card-title">Expenses</v-card-title>

      <v-card-text>
        <ServerSideTable :api-u-r-l="`${API_BASE}/expenses`" :headers="headers" root-key="expenses" />
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { API_BASE } from '@/config'

  const headers = [
    { title: 'ID', key: 'id', align: 'start' },
    { title: 'Description', key: 'description', align: 'start' },
    { title: 'Category', key: 'category', align: 'start' },
    { title: 'Amount', key: 'amount', align: 'end' },
    { title: 'Date', key: 'date', align: 'start' },
  ]
</script>
