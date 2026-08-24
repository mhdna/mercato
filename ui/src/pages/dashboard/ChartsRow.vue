<template>
  <v-card class="d-flex" flat style="gap: 0">
    <ChartCard :option="charts.revenue" title="Revenue" />
    <ChartCard :option="charts.expenses" title="Expenses" />
    <ChartCard :option="charts.profit" title="Profit" />
    <ChartCard :option="charts.margin" title="Profit Margin %" />
  </v-card>
</template>

<script setup>
  import ChartCard from './ChartCard.vue'
  import { buildChartOptions, generateMockData } from './chartOptions'

  const { dates, revenue, expenses } = generateMockData()
  const charts = buildChartOptions(dates, revenue, expenses)
</script>
