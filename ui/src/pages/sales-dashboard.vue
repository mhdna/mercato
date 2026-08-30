<template>
  <div class="sales-dashboard d-flex flex-column fill-height" style="overflow-y: auto">
    <!-- Header -->
    <div class="d-flex align-center flex-wrap ga-3 pa-4 pb-2">
      <div class="text-h6 font-weight-medium">Sales Dashboard</div>
      <v-spacer />
      <v-select
        v-model="selectedYear"
        density="compact"
        hide-details
        :items="years"
        label="Year"
        style="max-width: 120px"
        variant="outlined"
      />
      <v-select
        v-model="selectedMonth"
        density="compact"
        hide-details
        :items="['All', ...months]"
        label="Month"
        style="max-width: 130px"
        variant="outlined"
      />
    </div>

    <v-container class="pa-4 pt-2" fluid>
      <!-- Row 1: headline KPIs -->
      <v-row>
        <v-col
          v-for="kpi in kpis"
          :key="kpi.label"
          cols="12"
          md="4"
          sm="6"
        >
          <v-card class="pa-4 d-flex align-center justify-space-between" variant="tonal">
            <div>
              <div class="text-caption text-uppercase text-medium-emphasis">{{ kpi.label }}</div>
              <div class="text-h4 font-weight-medium mt-1">{{ kpi.value }}</div>
            </div>
            <v-icon color="primary" :icon="kpi.icon" size="40" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Row 2: monthly trend + highlight cards -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Monthly — Sales, Profit &amp; Profit %</div>
            <v-chart autoresize class="chart chart--tall" :option="monthlyOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <div class="d-flex flex-column ga-3 fill-height">
            <v-card class="pa-4 flex-grow-1 d-flex flex-column justify-center" variant="flat">
              <div class="chart-card__title">Top Product</div>
              <div class="text-h5 font-weight-medium mt-2">{{ topProduct.name }}</div>
              <div class="d-flex ga-6 mt-2">
                <div>
                  <div class="text-caption text-medium-emphasis">Quantity</div>
                  <div class="text-subtitle-1 font-weight-medium">{{ topProduct.qty }}</div>
                </div>
                <div>
                  <div class="text-caption text-medium-emphasis">Revenue</div>
                  <div class="text-subtitle-1 font-weight-medium">{{ money(topProduct.value) }}</div>
                </div>
              </div>
            </v-card>
            <v-card class="pa-4 flex-grow-1 d-flex flex-column justify-center" variant="flat">
              <div class="chart-card__title">Top Category</div>
              <div class="text-h5 font-weight-medium mt-2">{{ topCategory.name }}</div>
              <div class="mt-2">
                <div class="text-caption text-medium-emphasis">Revenue</div>
                <div class="text-subtitle-1 font-weight-medium">{{ money(topCategory.value) }}</div>
              </div>
            </v-card>
          </div>
        </v-col>
      </v-row>

      <!-- Row 3: product ranking + two splits -->
      <v-row>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Revenue by Product</div>
            <v-chart autoresize class="chart chart--tall" :option="productOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Sales Type</div>
            <v-chart autoresize class="chart chart--tall" :option="salesTypeOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Payment Mode</div>
            <v-chart autoresize class="chart chart--tall" :option="paymentOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Row 4: daily trend + category treemap -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Daily Sales</div>
            <v-chart autoresize class="chart chart--tall" :option="dailyOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Category</div>
            <v-chart autoresize class="chart chart--tall" :option="categoryOption" />
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

  // Minimal palette: one accent (primary), a lighter tint for the
  // secondary series, neutral grey for structure. No decorative colour.
  const primary = computed(() => c.value.primary)
  const isDark = computed(() => theme.current.value.dark)
  const tint = computed(() => primary.value + (isDark.value ? '77' : '55'))
  const axisColor = computed(() => (isDark.value ? '#3a3a3a' : '#e0e0e0'))
  const textColor = computed(() => (isDark.value ? '#aaa' : '#666'))

  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const years = ['2021', '2022']
  const selectedYear = ref('2022')
  const selectedMonth = ref('All')

  const money = v => '$' + Math.round(v).toLocaleString('en-US')

  // ---- Fake data ---------------------------------------------------------
  const kpis = [
    { label: 'Total Sales', value: '$401,412', icon: 'mdi-cash-multiple' },
    { label: 'Total Profit', value: '$68,908', icon: 'mdi-chart-box-outline' },
    { label: 'Profit %', value: '21%', icon: 'mdi-percent-outline' },
  ]

  const topProduct = { name: 'Product 41', qty: '132 Ft', value: 22_952 }
  const topCategory = { name: 'Category 04', value: 95_269 }

  const monthlySales = [38_000, 42_000, 31_000, 28_000, 33_000, 30_000, 34_000, 37_000, 45_000, 40_000, 44_000, 38_000]
  const profitPct = [21, 22, 22, 25, 17, 23, 18, 19, 23, 20, 23, 19]
  const monthlyProfit = monthlySales.map((s, i) => Math.round((s * profitPct[i]) / 100))

  const products = [
    { name: 'Product 11', value: 5856 },
    { name: 'Product 12', value: 11_583 },
    { name: 'Product 13', value: 8424 },
    { name: 'Product 14', value: 12_765 },
    { name: 'Product 15', value: 1839 },
    { name: 'Product 16', value: 1997 },
    { name: 'Product 17', value: 9877 },
    { name: 'Product 18', value: 4035 },
    { name: 'Product 19', value: 20_160 },
    { name: 'Product 20', value: 8006 },
  ]

  const salesType = [
    { name: 'Direct Sales', value: 52 },
    { name: 'Online', value: 33 },
    { name: 'Wholesaler', value: 15 },
  ]

  const paymentMode = [
    { name: 'Cash', value: 50 },
    { name: 'Online', value: 50 },
  ]

  const daily = [
    8200, 15_400, 6100, 11_800, 9300, 18_600, 7400, 21_200, 12_900, 5600,
    16_700, 10_300, 8800, 22_400, 14_100, 6900, 19_800, 11_200, 9600, 24_100,
    13_300, 7700, 17_900, 10_800, 8100, 20_600, 12_400, 6300, 15_900, 11_500, 9100,
  ]

  const categories = [
    { name: 'Category 04', value: 95_269 },
    { name: 'Category 02', value: 92_964 },
    { name: 'Category 05', value: 91_617 },
    { name: 'Category 01', value: 69_262 },
    { name: 'Category 03', value: 52_300 },
  ]

  // ---- Shared axis styling ---------------------------------------------
  const catAxis = computed(() => ({
    axisLine: { lineStyle: { color: axisColor.value } },
    axisTick: { show: false },
    axisLabel: { color: textColor.value, fontSize: 11 },
  }))
  const valAxis = computed(() => ({
    axisLabel: { color: textColor.value, fontSize: 11 },
    splitLine: { lineStyle: { color: axisColor.value, type: 'dashed' } },
  }))

  // ---- Chart options -------------------------------------------------
  const monthlyOption = computed(() => ({
    grid: { top: 32, right: 48, bottom: 28, left: 56 },
    tooltip: { trigger: 'axis', appendTo: 'body' },
    legend: {
      top: 0,
      right: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: textColor.value, fontSize: 10 },
    },
    xAxis: { type: 'category', data: months, ...catAxis.value },
    yAxis: [
      { type: 'value', axisLabel: { color: textColor.value, fontSize: 11, formatter: v => '$' + v / 1000 + 'k' }, splitLine: { lineStyle: { color: axisColor.value, type: 'dashed' } } },
      { type: 'value', max: 40, axisLabel: { color: textColor.value, fontSize: 11, formatter: '{value}%' }, splitLine: { show: false } },
    ],
    series: [
      { name: 'Sales', type: 'bar', barWidth: '32%', itemStyle: { color: primary.value }, data: monthlySales },
      { name: 'Profit', type: 'bar', barWidth: '32%', itemStyle: { color: tint.value }, data: monthlyProfit },
      {
        name: 'Profit %',
        type: 'line',
        yAxisIndex: 1,
        smooth: true,
        symbol: 'circle',
        symbolSize: 5,
        lineStyle: { width: 2, color: textColor.value },
        itemStyle: { color: textColor.value },
        data: profitPct,
      },
    ],
  }))

  const productOption = computed(() => {
    const sorted = products.toSorted((a, b) => a.value - b.value)
    return {
      grid: { top: 8, right: 56, bottom: 8, left: 72 },
      tooltip: { trigger: 'axis', appendTo: 'body', valueFormatter: money },
      xAxis: { type: 'value', ...valAxis.value, axisLabel: { show: false }, splitLine: { show: false } },
      yAxis: { type: 'category', data: sorted.map(p => p.name), ...catAxis.value },
      series: [
        {
          type: 'bar',
          barWidth: 12,
          itemStyle: { color: primary.value, borderRadius: [0, 3, 3, 0] },
          label: { show: true, position: 'right', color: textColor.value, fontSize: 10, formatter: p => money(p.value) },
          data: sorted.map(p => p.value),
        },
      ],
    }
  })

  function donut (data, center = ['50%', '52%']) {
    return {
      tooltip: { trigger: 'item', appendTo: 'body', formatter: '{b}: {d}%' },
      legend: { bottom: 0, itemWidth: 10, itemHeight: 10, textStyle: { color: textColor.value, fontSize: 10 } },
      series: [
        {
          type: 'pie',
          radius: ['48%', '72%'],
          center,
          avoidLabelOverlap: true,
          label: { show: true, formatter: '{d}%', color: textColor.value, fontSize: 11 },
          labelLine: { length: 8, length2: 6 },
          data: data.map((d, i) => ({
            name: d.name,
            value: d.value,
            itemStyle: { color: primary.value + ['ff', 'aa', '66', '40'][i % 4] },
          })),
        },
      ],
    }
  }

  const salesTypeOption = computed(() => donut(salesType))
  const paymentOption = computed(() => donut(paymentMode))

  const dailyOption = computed(() => ({
    grid: { top: 16, right: 16, bottom: 28, left: 56 },
    tooltip: { trigger: 'axis', appendTo: 'body', valueFormatter: money },
    xAxis: { type: 'category', boundaryGap: false, data: daily.map((_, i) => i + 1), ...catAxis.value },
    yAxis: { type: 'value', axisLabel: { color: textColor.value, fontSize: 11, formatter: v => '$' + v / 1000 + 'k' }, splitLine: { lineStyle: { color: axisColor.value, type: 'dashed' } } },
    series: [
      {
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 2, color: primary.value },
        areaStyle: { color: tint.value, opacity: 0.5 },
        data: daily,
      },
    ],
  }))

  const categoryOption = computed(() => ({
    tooltip: { appendTo: 'body', formatter: p => `${p.name}<br/>${money(p.value)}` },
    series: [
      {
        type: 'treemap',
        roam: false,
        nodeClick: false,
        breadcrumb: { show: false },
        itemStyle: { borderColor: isDark.value ? '#000' : '#fff', borderWidth: 2, gapWidth: 2 },
        label: { formatter: p => `${p.name}\n${p.value.toLocaleString('en-US')}`, fontSize: 11, color: '#fff' },
        data: categories.map((cat, i) => ({
          name: cat.name,
          value: cat.value,
          itemStyle: { color: primary.value + ['ff', 'd0', 'a8', '80', '58'][i % 5] },
        })),
      },
    ],
  }))
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

.chart {
  height: 240px;
  width: 100%;
}

.chart--tall {
  height: 300px;
}
</style>
