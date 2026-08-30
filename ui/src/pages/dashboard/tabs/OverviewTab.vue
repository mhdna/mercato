<template>
  <div class="d-flex flex-column flex-grow-1" style="height: 100%; min-height: 0; overflow: hidden">
    <v-alert
      v-if="error"
      class="ma-2"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ error }}
    </v-alert>
    <div class="kpi-grid mt-4 mb-3">
      <SparklineCard
        icon="mdi-currency-usd"
        :loading="loading"
        :prev-value="prevMetric('revenue')"
        :series="revenueSeries"
        show-axis
        title="Revenue"
        unit="$"
        :value="metric('revenue')"
      />
      <SparklineCard
        icon="mdi-currency-usd-off"
        :loading="loading"
        :prev-value="prevMetric('expenses')"
        :series="expensesSeries"
        show-axis
        title="Expenses"
        unit="$"
        :value="metric('expenses')"
      />
      <SparklineCard
        icon="mdi-trending-up"
        :loading="loading"
        :prev-value="prevProfit"
        :series="profitSeries"
        show-axis
        title="Profit"
        unit="$"
        :value="profit"
      />
      <SparklineCard
        icon="mdi-percent-outline"
        :loading="loading"
        :prev-value="prevMargin"
        :series="marginSeries"
        show-axis
        title="Profit Margin"
        unit="%"
        :value="margin"
      />
      <SparklineCard
        icon="mdi-scale-balance"
        :loading="loading"
        :prev-value="prevFinPosition"
        :series="finPositionSeries"
        show-axis
        title="Financial Position"
        unit="$"
        :value="finPosition"
      />
    </div>
    <!-- ECharts trend row -- superseded by the sparkline KPI tiles above. -->
    <!-- <ChartsRow class="mb-6" :series="series" /> -->
    <TablesRow style="flex: 1 1 0; min-height: 320px" />
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import SparklineCard from '@/components/Cards/SparklineCard.vue'
  import { useDashboardSummary } from '@/composables/useDashboardSummary'
  // import ChartsRow from '../ChartsRow.vue'
  import TablesRow from '../TablesRow.vue'

  const { summary, prevSummary, loading, error } = useDashboardSummary()

  function metric (key) {
    const v = summary.value?.[key]
    return typeof v === 'number' ? v : null
  }

  function prevMetric (key) {
    const v = prevSummary.value?.[key]
    return typeof v === 'number' ? v : null
  }

  const profit = computed(() => {
    const r = metric('revenue')
    const e = metric('expenses')
    return r == null || e == null ? null : r - e
  })

  const prevProfit = computed(() => {
    const r = prevMetric('revenue')
    const e = prevMetric('expenses')
    return r == null || e == null ? null : r - e
  })

  // Profit as a percentage of revenue, rounded to one decimal.
  function marginOf (rev, prof) {
    if (rev == null || prof == null || rev === 0) return null
    return Math.round((prof / rev) * 1000) / 10
  }
  const margin = computed(() => marginOf(metric('revenue'), profit.value))
  const prevMargin = computed(() => marginOf(prevMetric('revenue'), prevProfit.value))

  // Financial position = revenue - expenses - purchases, in cents. Purchases
  // are admin-only, so they're null (treated as 0) under a branch view.
  function netOf (m) {
    const r = m('revenue')
    const e = m('expenses')
    if (r == null || e == null) return null
    return r - e - (m('purchases') ?? 0)
  }
  const finPosition = computed(() => netOf(metric))
  const prevFinPosition = computed(() => netOf(prevMetric))

  // Per-day values (raw cents) driving each KPI tile's trailing sparkline.
  const daily = computed(() => summary.value?.series ?? [])
  const revenueSeries = computed(() => daily.value.map(d => d.revenue ?? 0))
  const expensesSeries = computed(() => daily.value.map(d => d.expenses ?? 0))
  const profitSeries = computed(() => daily.value.map(d => (d.revenue ?? 0) - (d.expenses ?? 0)))
  const marginSeries = computed(() => daily.value.map(d => {
    const rev = d.revenue ?? 0
    return rev === 0 ? 0 : Math.round((((d.revenue ?? 0) - (d.expenses ?? 0)) / rev) * 1000) / 10
  }))
  // Running total of daily net, so the tile shows the trajectory.
  const finPositionSeries = computed(() => {
    let running = 0
    return daily.value.map(d => {
      running += (d.revenue ?? 0) - (d.expenses ?? 0) - (d.purchases ?? 0)
      return running
    })
  })
</script>

<style scoped>
/* The four headline financial KPIs -- always one tight row. */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

</style>
