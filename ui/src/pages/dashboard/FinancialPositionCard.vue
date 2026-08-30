<template>
  <v-card class="fin-position pa-4" flat rounded="0">
    <div class="d-flex flex-wrap align-center ga-6">
      <!-- Headline + state -->
      <div style="flex: 1 1 220px; min-width: 0">
        <div class="d-flex align-center text-subtitle-2 text-uppercase text-medium-emphasis mb-1">
          <v-icon class="me-2" color="indigo-lighten-2" :icon="stateIcon" size="20" />
          Financial Position
        </div>

        <v-progress-circular
          v-if="loading && summary == null"
          color="indigo-lighten-2"
          indeterminate
          size="26"
          width="2"
        />

        <template v-else>
          <div class="d-flex align-baseline ga-3">
            <span class="text-h4 font-weight-black fin-figure" :class="`text-${stateColor}`">
              {{ headline }}
            </span>
            <v-chip :color="stateColor" label size="small" variant="tonal">
              <v-icon :icon="stateIcon" start />
              {{ stateLabel }}
            </v-chip>
          </div>

          <div class="text-caption mt-1" :class="deltaClass">
            <v-icon :icon="deltaIcon" size="14" />
            {{ deltaText }}
          </div>
        </template>
      </div>

      <!-- Breakdown -->
      <div class="d-flex flex-wrap ga-4 fin-breakdown" style="flex: 1 1 260px">
        <div v-for="part in breakdown" :key="part.label">
          <div class="text-caption text-medium-emphasis">{{ part.label }}</div>
          <div class="text-body-1 font-weight-medium fin-figure">{{ part.value }}</div>
        </div>
      </div>

      <!-- Cumulative-net trajectory -->
      <div style="flex: 1 1 240px; min-width: 200px">
        <VChart autoresize :option="sparkOption" style="height: 96px; width: 100%" />
      </div>
    </div>
  </v-card>
</template>

<script setup>
  import { computed } from 'vue'
  import VChart from 'vue-echarts'
  import { formatCompactMoney, formatMoney } from '@/utils/money'

  const props = defineProps({
    // The /dashboard/summary payload for the selected period (or null while
    // first loading).
    summary: { type: Object, default: null },
    // Same shape for the immediately-preceding period, for the trend line.
    prevSummary: { type: Object, default: null },
    loading: { type: Boolean, default: false },
  })

  // revenue - expenses - purchases, in cents. Purchases are admin-only, so
  // they're 0 under a branch-scoped view -- the number still holds.
  function net (s) {
    if (s == null) return null
    return (s.revenue ?? 0) - (s.expenses ?? 0) - (s.purchases ?? 0)
  }

  const current = computed(() => net(props.summary))
  const previous = computed(() => net(props.prevSummary))

  const isLoss = computed(() => (current.value ?? 0) < 0)
  const stateColor = computed(() => (isLoss.value ? 'error' : 'success'))
  const stateLabel = computed(() => (isLoss.value ? 'Losing' : 'Gaining'))
  const stateIcon = computed(() => (isLoss.value ? 'mdi-trending-down' : 'mdi-trending-up'))

  const headline = computed(() =>
    current.value == null ? '—' : formatCompactMoney(current.value),
  )

  // Period-over-period change in the net. Undefined when there's no prior
  // baseline to divide by.
  const deltaPct = computed(() => {
    if (current.value == null || previous.value == null || previous.value === 0) return null
    return ((current.value - previous.value) / Math.abs(previous.value)) * 100
  })
  const deltaUp = computed(
    () => current.value != null && previous.value != null && current.value >= previous.value,
  )
  const deltaIcon = computed(() => (deltaUp.value ? 'mdi-arrow-up' : 'mdi-arrow-down'))
  const deltaClass = computed(() => (deltaUp.value ? 'text-success' : 'text-error'))
  const deltaText = computed(() => {
    if (deltaPct.value == null) return 'No prior period to compare'
    const sign = deltaPct.value >= 0 ? '+' : '−'
    return `${sign}${Math.abs(deltaPct.value).toFixed(1)}% vs. previous period`
  })

  const marginPct = computed(() => {
    const rev = props.summary?.revenue ?? 0
    if (!rev || current.value == null) return null
    return (current.value / rev) * 100
  })

  const breakdown = computed(() => {
    const s = props.summary ?? {}
    return [
      { label: 'Revenue', value: `$${formatMoney(s.revenue ?? 0)}` },
      { label: 'Expenses', value: `$${formatMoney(s.expenses ?? 0)}` },
      { label: 'Purchases', value: `$${formatMoney(s.purchases ?? 0)}` },
      {
        label: 'Margin',
        value: marginPct.value == null ? '—' : `${marginPct.value.toFixed(1)}%`,
      },
    ]
  })

  // Running total of daily (revenue - expenses - purchases), in dollars, so
  // the reader sees the trajectory and where it crosses zero.
  const cumulative = computed(() => {
    const series = props.summary?.series ?? []
    let running = 0
    return series.map(d => {
      running += ((d.revenue ?? 0) - (d.expenses ?? 0) - (d.purchases ?? 0)) / 100
      return Math.round(running)
    })
  })

  const sparkOption = computed(() => {
    const data = cumulative.value
    const days = (props.summary?.series ?? []).map(d => d.day)
    const color = isLoss.value ? '#EA4335' : '#34A853'
    return {
      grid: { top: 6, right: 6, bottom: 6, left: 6, containLabel: false },
      xAxis: { type: 'category', data: days, show: false, boundaryGap: false },
      yAxis: { type: 'value', show: false },
      tooltip: {
        trigger: 'axis',
        valueFormatter: v => `$${Number(v).toLocaleString()}`,
        appendTo: 'body',
      },
      series: [
        {
          name: 'Cumulative net',
          type: 'line',
          smooth: true,
          symbol: 'none',
          lineStyle: { width: 2, color },
          areaStyle: { color: isLoss.value ? 'rgba(234,67,53,0.14)' : 'rgba(52,168,83,0.14)' },
          data,
          markLine: {
            silent: true,
            symbol: 'none',
            label: { show: false },
            lineStyle: { color: '#888', type: 'dashed', width: 1 },
            data: [{ yAxis: 0 }],
          },
        },
      ],
    }
  })
</script>

<style scoped>
  .fin-position {
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  }

  /* Fixed-width digits so figures don't jitter between filter changes. */
  .fin-figure {
    font-variant-numeric: tabular-nums;
  }
</style>
