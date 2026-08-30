<template>
  <v-card
    class="summary-card"
    rounded="lg"
    variant="outlined"
    width="100%"
  >
    <div class="summary-card-body">
      <div class="text-body-2 text-medium-emphasis text-uppercase">
        {{ props.title }}
      </div>

      <!-- Only the very first load (no value yet) shows a spinner. Once a
        number is on screen it stays put through every background refresh --
        the value just changes in place, no flip to a loader, no reflow. -->
      <v-progress-circular
        v-if="props.loading && props.value == null"
        class="mt-3"
        color="indigo-lighten-2"
        indeterminate
        size="22"
        width="2"
      />
      <div v-else class="d-flex align-center flex-wrap ga-2 mt-2">
        <span class="metric-value text-h4 font-weight-black">
          <strong v-if="props.unit === '$' && display !== '—'" class="me-1">{{ props.unit }}</strong>{{ display }}
        </span>
        <v-chip
          v-if="deltaPct != null"
          :color="deltaUp ? 'success' : 'error'"
          label
          size="small"
          variant="tonal"
        >
          <v-icon :icon="deltaUp ? 'mdi-arrow-up' : 'mdi-arrow-down'" size="14" start />
          {{ deltaText }}
        </v-chip>
      </div>
    </div>

    <VChart
      v-if="props.series.length > 1"
      autoresize
      class="summary-card-spark"
      :option="option"
    />
  </v-card>
</template>
<script setup>
  import { computed } from 'vue'
  import VChart from 'vue-echarts'
  import { formatMoney } from '@/utils/money'

  const props = defineProps({
    title: {
      type: String,
      required: true,
    },
    // Kept for call-site compatibility; the image-style card has no icon.
    icon: {
      type: String,
      default: '',
    },
    unit: {
      type: String,
      required: true,
    },
    // Raw metric value from /dashboard/summary. Money metrics ($ unit)
    // arrive in cents. null renders as an em dash.
    value: {
      type: Number,
      default: null,
    },
    // Same metric for the immediately-preceding period, for the delta chip.
    prevValue: {
      type: Number,
      default: null,
    },
    // Per-day values for the trailing sparkline (same unit as `value`).
    series: {
      type: Array,
      default: () => [],
    },
    showAxis: {
      type: Boolean,
      default: false,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  })

  const display = computed(() => {
    if (props.value == null || Number.isNaN(props.value)) return '—'
    if (props.unit === '$') return formatMoney(props.value)
    if (props.unit === '%') return `${props.value.toLocaleString()}%`
    return props.value.toLocaleString()
  })

  // Period-over-period change. Null when there's no prior baseline to divide by.
  const deltaPct = computed(() => {
    if (props.value == null || props.prevValue == null || props.prevValue === 0) return null
    return ((props.value - props.prevValue) / Math.abs(props.prevValue)) * 100
  })
  const deltaUp = computed(() => (deltaPct.value ?? 0) >= 0)
  const deltaText = computed(() => {
    if (deltaPct.value == null) return ''
    const sign = deltaPct.value >= 0 ? '+' : '−'
    return `${sign}${Math.abs(deltaPct.value).toFixed(2)}%`
  })

  // Filled area sparkline that bleeds to the card edges -- same look as the
  // Vuetify v-sparkline it replaces, just drawn by ECharts so it matches the
  // rest of the dashboard's charting.
  function fmtTip (v) {
    if (v == null || Number.isNaN(v)) return '—'
    if (props.unit === '$') return `$${formatMoney(v)}`
    if (props.unit === '%') return `${Number(v).toLocaleString()}%`
    return Number(v).toLocaleString()
  }

  const option = computed(() => {
    const base = deltaUp.value ? '76, 175, 80' : '244, 67, 54'
    return {
      animation: true,
      grid: { left: 0, right: 0, top: 8, bottom: 0, containLabel: false },
      tooltip: {
        trigger: 'axis',
        appendTo: 'body',
        axisPointer: { type: 'line', lineStyle: { color: `rgb(${base})`, width: 1 } },
        formatter: params => fmtTip(params?.[0]?.value),
      },
      xAxis: {
        type: 'category',
        show: props.showAxis,
        boundaryGap: false,
        data: props.series.map((_, i) => i),
        axisLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.45)' } },
        axisTick: { show: props.showAxis },
        axisLabel: { show: false },
      },
      yAxis: {
        type: 'value',
        show: props.showAxis,
        scale: true,
        axisLine: { show: true, lineStyle: { color: 'rgba(128, 128, 128, 0.45)' } },
        axisTick: { show: props.showAxis },
        axisLabel: { show: false },
        splitLine: { show: props.showAxis, lineStyle: { color: 'rgba(128, 128, 128, 0.18)' } },
      },
      series: [
        {
          type: 'line',
          data: props.series,
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          showSymbol: false,
          lineStyle: { width: 1.5, color: `rgb(${base})`, cap: 'round' },
          itemStyle: { color: `rgb(${base})` },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: `rgba(${base}, 0.55)` },
                { offset: 1, color: `rgba(${base}, 0)` },
              ],
            },
          },
        },
      ],
    }
  })
</script>

<style scoped>
  .summary-card {
    position: relative;
    overflow: hidden;
    height: 232px;
    min-height: 232px;
  }

  .summary-card-body {
    position: relative;
    z-index: 1;
    box-sizing: border-box;
    height: 104px;
    overflow: hidden;
    padding: 18px 20px;
  }

  /* Bleed the sparkline to the card edges along the bottom, like the KPI
     tiles it's modelled on. */
  .summary-card-spark {
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 128px;
  }

  /* Fixed-width digits so the number doesn't jitter sideways as it
     changes between filter/period selections. */
  .metric-value {
    font-variant-numeric: tabular-nums;
  }
</style>
