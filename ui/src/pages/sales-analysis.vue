<template>
  <div class="sales-analysis d-flex flex-column fill-height" style="overflow-y: auto">
    <!-- Header -->
    <div class="d-flex align-center flex-wrap ga-3 pa-4 pb-2">
      <div class="text-h6 font-weight-medium">Sales Analysis</div>
      <v-spacer />
      <div class="text-caption text-medium-emphasis">YTD through</div>
      <v-select
        v-model="selectedMonth"
        density="compact"
        hide-details
        :items="months"
        style="max-width: 130px"
        variant="outlined"
      />
    </div>

    <v-container class="pa-4 pt-2" fluid>
      <!-- Row 1: KPIs + two charts -->
      <v-row>
        <v-col cols="12" md="3">
          <div class="d-flex flex-column ga-3 fill-height">
            <v-card
              v-for="kpi in kpis"
              :key="kpi.label"
              class="pa-3 d-flex flex-column justify-center flex-grow-1"
              variant="tonal"
            >
              <div class="d-flex align-center justify-space-between">
                <div class="text-h5 font-weight-medium">{{ kpi.value }}</div>
                <v-progress-circular
                  v-if="kpi.progress != null"
                  :color="kpi.progress >= 100 ? 'primary' : 'error'"
                  :model-value="Math.min(kpi.progress, 100)"
                  :size="44"
                  :width="4"
                >
                  <span class="text-caption">{{ kpi.progress }}%</span>
                </v-progress-circular>
                <v-chip
                  v-else-if="kpi.delta != null"
                  :color="kpi.delta >= 0 ? 'primary' : 'error'"
                  size="small"
                  variant="tonal"
                >
                  <v-icon :icon="kpi.delta >= 0 ? 'mdi-arrow-up' : 'mdi-arrow-down'" start />
                  {{ Math.abs(kpi.delta) }}%
                </v-chip>
              </div>
              <div class="text-caption text-medium-emphasis mt-1">{{ kpi.label }}</div>
            </v-card>
          </div>
        </v-col>

        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Sales by Product / Category</div>
            <v-chart autoresize class="chart" :option="productOption" />
          </v-card>
        </v-col>

        <v-col cols="12" md="5">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Actual Sales vs Target</div>
            <div class="chart-card__subtitle">Bar colour reflects attainment vs. monthly target</div>
            <v-chart autoresize class="chart" :option="actualVsTargetOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Row 2 -->
      <v-row>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Sales by Region</div>
            <v-chart autoresize class="chart" :option="regionOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Sales by Channel</div>
            <v-chart autoresize class="chart" :option="channelOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Inventory Days Outstanding</div>
            <div class="chart-card__subtitle">
              Target {{ inventoryDays.target }} · Last year {{ inventoryDays.lastYear }}
            </div>
            <v-chart autoresize class="chart" :option="inventoryGaugeOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Row 3 -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Sales Growth</div>
            <div class="chart-card__subtitle">Month-over-month change</div>
            <v-chart autoresize class="chart" :option="growthOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Receivable Days Outstanding</div>
            <div class="chart-card__subtitle">
              Target {{ receivableDays.target }} · Last year {{ receivableDays.lastYear }}
            </div>
            <v-chart autoresize class="chart" :option="receivableGaugeOption" />
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import VChart from 'vue-echarts'
  import { useTheme } from 'vuetify'

  const theme = useTheme()
  const c = computed(() => theme.current.value.colors)

  // Palette — deliberately minimal: one accent (primary), a neutral grey
  // for "target" / reference marks, red only for shortfalls and declines.
  const primary = computed(() => c.value.primary)
  const negative = computed(() => c.value.error)
  const isDark = computed(() => theme.current.value.dark)
  const muted = computed(() => (isDark.value ? '#5c5c5c' : '#bdbdbd'))
  const axisColor = computed(() => (isDark.value ? '#3a3a3a' : '#e0e0e0'))
  const textColor = computed(() => (isDark.value ? '#aaa' : '#666'))

  const months = ['JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC', 'JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN']
  const selectedMonth = ref('JUN')

  // ---- Fake data ------------------------------------------------------------
  const kpis = [
    { label: 'Sales', value: '16,461', delta: 43 },
    { label: 'Sales Target', value: '15,907' },
    { label: 'Target Achievement', value: '', progress: 103 },
    { label: 'Sales Last Year', value: '11,545' },
    { label: 'Gross Profit', value: '5,868' },
    { label: 'Gross Profit Margin', value: '36%' },
  ]

  const products = [
    { name: 'Product 1', actual: 5597, target: 5400 },
    { name: 'Product 2', actual: 4263, target: 4100 },
    { name: 'Product 3', actual: 3405, target: 3550 },
    { name: 'Product 4', actual: 2379, target: 2650 },
    { name: 'Product 5', actual: 817, target: 760 },
  ]

  const monthlyActual = [1210, 1550, 1676, 1527, 962, 957, 1190, 1778, 1388, 1450, 1550, 1223]
  const monthlyTarget = [1200, 1750, 1650, 1560, 1300, 1300, 1250, 1560, 1420, 1450, 1720, 1260]

  const regions = [
    { name: 'East', actual: 5761, target: 5567 },
    { name: 'West', actual: 4115, target: 3977 },
    { name: 'South', actual: 4938, target: 4772 },
    { name: 'North', actual: 1646, target: 1591 },
  ]

  const channels = [
    { name: 'Distributor', value: 5761 },
    { name: 'Dealer', value: 4115 },
    { name: 'Retail', value: 3292 },
    { name: 'Corporate', value: 2469 },
    { name: 'Online', value: 823 },
  ]

  const growth = [5, 28, 8, -9, -37, -1, 24, 49, -22, 4, 7, -21]

  const inventoryDays = { value: 42, target: 33, lastYear: 36, max: 180 }
  const receivableDays = { value: 32, target: 30, lastYear: 33, max: 180 }

  // ---- Shared axis styling ------------------------------------------------
  const catAxis = computed(() => ({
    axisLine: { lineStyle: { color: axisColor.value } },
    axisTick: { show: false },
    axisLabel: { color: textColor.value, fontSize: 11 },
  }))
  const valAxis = computed(() => ({
    axisLabel: { color: textColor.value, fontSize: 11 },
    splitLine: { lineStyle: { color: axisColor.value, type: 'dashed' } },
  }))

  // ---- Chart options ----------------------------------------------------
  const productOption = computed(() => ({
    grid: { top: 28, right: 16, bottom: 24, left: 48 },
    tooltip: { trigger: 'axis', appendTo: 'body' },
    legend: {
      top: 0,
      right: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: textColor.value, fontSize: 10 },
      data: ['Sales', 'Target'],
    },
    xAxis: { type: 'category', data: products.map(p => p.name), ...catAxis.value },
    yAxis: { type: 'value', ...valAxis.value },
    series: [
      {
        name: 'Sales',
        type: 'bar',
        barWidth: '46%',
        label: { show: true, position: 'top', color: textColor.value, fontSize: 10 },
        data: products.map(p => ({
          value: p.actual,
          itemStyle: { color: p.actual >= p.target ? primary.value : muted.value },
        })),
      },
      {
        name: 'Target',
        type: 'scatter',
        symbol: 'rect',
        symbolSize: [30, 3],
        itemStyle: { color: textColor.value },
        data: products.map(p => p.target),
      },
    ],
  }))

  const actualVsTargetOption = computed(() => {
    const barColor = i => {
      const ratio = monthlyActual[i] / monthlyTarget[i]
      if (ratio >= 1) return primary.value
      if (ratio >= 0.9) return muted.value
      return negative.value
    }
    return {
      grid: { top: 16, right: 12, bottom: 24, left: 44 },
      tooltip: { trigger: 'axis', appendTo: 'body' },
      xAxis: { type: 'category', data: months, ...catAxis.value },
      yAxis: { type: 'value', ...valAxis.value },
      series: [
        {
          name: 'Actual',
          type: 'bar',
          barWidth: '52%',
          data: monthlyActual.map((v, i) => ({ value: v, itemStyle: { color: barColor(i) } })),
        },
        {
          name: 'Target',
          type: 'scatter',
          symbol: 'rect',
          symbolSize: [22, 3],
          itemStyle: { color: textColor.value },
          data: monthlyTarget,
        },
      ],
    }
  })

  const regionOption = computed(() => ({
    grid: { top: 24, right: 24, bottom: 16, left: 52 },
    tooltip: { trigger: 'axis', appendTo: 'body' },
    legend: {
      top: 0,
      right: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: textColor.value, fontSize: 10 },
      data: ['Actual', 'Target'],
    },
    xAxis: { type: 'value', ...valAxis.value },
    yAxis: { type: 'category', data: regions.map(r => r.name), inverse: true, ...catAxis.value },
    series: [
      {
        name: 'Actual',
        type: 'bar',
        barWidth: 10,
        itemStyle: { color: primary.value },
        data: regions.map(r => r.actual),
      },
      {
        name: 'Target',
        type: 'bar',
        barWidth: 10,
        itemStyle: { color: muted.value },
        data: regions.map(r => r.target),
      },
    ],
  }))

  const channelOption = computed(() => {
    const alphas = ['ff', 'cc', '99', '66', '40']
    return {
      tooltip: { trigger: 'item', appendTo: 'body', formatter: '{b}: {c} ({d}%)' },
      legend: {
        type: 'scroll',
        orient: 'vertical',
        right: 0,
        top: 'center',
        itemWidth: 10,
        itemHeight: 10,
        textStyle: { color: textColor.value, fontSize: 10 },
      },
      series: [
        {
          type: 'pie',
          radius: ['45%', '72%'],
          center: ['38%', '52%'],
          avoidLabelOverlap: true,
          label: { show: false },
          labelLine: { show: false },
          data: channels.map((ch, i) => ({
            name: ch.name,
            value: ch.value,
            itemStyle: { color: primary.value + alphas[i % alphas.length] },
          })),
        },
      ],
    }
  })

  const growthOption = computed(() => ({
    grid: { top: 24, right: 12, bottom: 24, left: 40 },
    tooltip: { trigger: 'axis', appendTo: 'body', valueFormatter: v => `${v}%` },
    xAxis: { type: 'category', data: months, ...catAxis.value },
    yAxis: {
      type: 'value',
      axisLabel: { color: textColor.value, fontSize: 11, formatter: '{value}%' },
      splitLine: { lineStyle: { color: axisColor.value, type: 'dashed' } },
    },
    series: [
      {
        type: 'bar',
        barWidth: '52%',
        label: {
          show: true,
          color: textColor.value,
          fontSize: 10,
          formatter: p => `${p.value}%`,
        },
        data: growth.map(v => ({
          value: v,
          label: { position: v >= 0 ? 'top' : 'bottom' },
          itemStyle: { color: v >= 0 ? primary.value : negative.value },
        })),
      },
    ],
  }))

  function gaugeOption (data, color) {
    return {
      series: [
        {
          type: 'gauge',
          min: 0,
          max: data.max,
          startAngle: 200,
          endAngle: -20,
          radius: '92%',
          center: ['50%', '62%'],
          progress: { show: true, width: 12, itemStyle: { color } },
          axisLine: { lineStyle: { width: 12, color: [[1, axisColor.value]] } },
          axisTick: { show: false },
          splitLine: { length: 8, lineStyle: { color: axisColor.value } },
          axisLabel: { show: false },
          pointer: { width: 4, itemStyle: { color } },
          anchor: { show: true, size: 8, itemStyle: { color } },
          title: { show: false },
          detail: {
            valueAnimation: true,
            offsetCenter: [0, '32%'],
            fontSize: 26,
            fontWeight: 500,
            color: textColor.value,
            formatter: '{value} days',
          },
          data: [{ value: data.value }],
        },
      ],
    }
  }

  const inventoryGaugeOption = computed(() => gaugeOption(inventoryDays, primary.value))
  const receivableGaugeOption = computed(() => gaugeOption(receivableDays, primary.value))
</script>

<style scoped>
.chart-card {
  height: 100%;
  padding: 12px;
  border: 1px solid rgba(128, 128, 128, 0.18);
}

.chart-card__title {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  opacity: 0.7;
}

.chart-card__subtitle {
  font-size: 0.7rem;
  opacity: 0.5;
  margin-bottom: 2px;
}

.chart {
  height: 240px;
  width: 100%;
}
</style>
