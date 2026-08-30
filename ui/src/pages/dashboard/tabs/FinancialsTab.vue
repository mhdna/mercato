<template>
  <div class="d-flex flex-column flex-grow-1" style="min-height: 0; overflow-y: auto">
    <v-card class="px-1 py-0 mb-2" flat>
      <div class="d-flex align-center justify-space-between pa-1 mb-n2">
        <v-card-title class="card-title pa-0">Revenue & Forecast</v-card-title>
        <v-btn-toggle
          v-model="zoomPreset"
          color="primary"
          density="compact"
          mandatory
          rounded="lg"
          variant="outlined"
          @update:model-value="initialView = false"
        >
          <v-btn value="day">Day</v-btn>
          <v-btn value="week">Week</v-btn>
          <v-btn value="month">Month</v-btn>
          <v-btn value="year">Year</v-btn>
        </v-btn-toggle>
      </div>
      <VChart autoresize :option="revenueForecastOption" style="height: 420px; width: 100%" />
      <div class="d-flex flex-wrap ga-4 px-2 pb-2 text-caption" style="opacity: 0.7">
        <span>🟩/🟥 Above/below forecast</span>
        <span>📌 Promotion</span>
        <span>♦️ Holiday</span>
        <span>▲ Stockout</span>
        <span>■ Unusually expensive day</span>
      </div>
    </v-card>

    <v-card class="d-flex" flat style="gap: 0">
      <FinancialChartCard :option="barOption" title="Revenue by Branch" />
      <FinancialChartCard :option="pieOption" title="Payment Methods" />
      <FinancialChartCard :option="marginOption" title="Profit Margin Trend" />
    </v-card>

    <v-card class="px-1 py-0 mt-2" flat>
      <v-card-title class="card-title pa-1 mb-n2">Sales Activity by Hour</v-card-title>
      <VChart autoresize :option="heatmapOption" style="height: 260px; width: 100%" />
    </v-card>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import VChart from 'vue-echarts'
  import { useSettingsStore } from '@/stores/settings'
  import FinancialChartCard from './FinancialChartCard.vue'
  import {
    buildHourlyHeatmapOption,
    buildMarginTrendOption,
    buildPaymentMethodPieOption,
    buildRevenueForecastOption,
    buildTopBranchesBarOption,
    DEFAULT_ZOOM_WINDOW,
    generateFinancialsSeries,
    ZOOM_PRESETS,
  } from './financialsChartOptions'

  const settingsStore = useSettingsStore()
  onMounted(() => settingsStore.init())

  const zoomPreset = ref('year')
  const initialView = ref(true)

  // Regenerated (new random walk) whenever the configured high season
  // months change, so flipping a month in Settings -> Financials is
  // visible here. Zoom presets just adjust the dataZoom window over the
  // same series, so they don't need to regenerate anything.
  const financialsSeries = computed(() => generateFinancialsSeries(undefined, settingsStore.financialsHighSeasonMonths))

  const revenueForecastOption = computed(() => buildRevenueForecastOption(
    financialsSeries.value,
    initialView.value ? DEFAULT_ZOOM_WINDOW : ZOOM_PRESETS[zoomPreset.value],
  ))

  const barOption = buildTopBranchesBarOption()
  const pieOption = buildPaymentMethodPieOption()
  const marginOption = buildMarginTrendOption()
  const heatmapOption = buildHourlyHeatmapOption()
</script>
