<template>
  <v-card class="d-flex" flat style="gap: 0">
    <ChartCard :option="charts.revenue" title="Revenue" />
    <ChartCard :option="charts.expenses" title="Expenses" />
    <ChartCard :option="charts.profit" title="Profit" />
    <ChartCard :option="charts.margin" title="Profit Margin %" />
  </v-card>
</template>

<script setup>
  import { computed } from 'vue'
  import ChartCard from './ChartCard.vue'
  import { buildChartOptions } from './chartOptions'

  const props = defineProps({
    // [{ day: 'YYYY-MM-DD', revenue: cents, expenses: cents }] from
    // /dashboard/summary.
    series: {
      type: Array,
      default: () => [],
    },
  })

  const charts = computed(() => {
    const dates = props.series.map(point =>
      new Date(`${point.day}T00:00:00`).toLocaleDateString('en-GB', { day: '2-digit', month: 'short' }),
    )
    // Charts show currency amounts, not cents.
    const revenue = props.series.map(point => Math.round((point.revenue ?? 0) / 100))
    const expenses = props.series.map(point => Math.round((point.expenses ?? 0) / 100))
    return buildChartOptions(dates, revenue, expenses)
  })
</script>
