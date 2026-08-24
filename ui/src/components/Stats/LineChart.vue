<template>
  <VChart
    autoresize
    class="chart"
    :loading="loading"
    :option="option"
    theme="light"
  />
</template>

<script setup>
  import { ref } from 'vue'
  import VChart from 'vue-echarts'
  import { useTheme } from 'vuetify'

  const theme = useTheme()
  const currentTheme = computed(() => theme.name)

  const loading = ref(false)

  const option = ref({
    tooltip: {
      trigger: 'axis',
      z: 9999,
      appendToBody: true,
    },
    grid: {
      top: '4%',
      left: '0%',
      right: '0%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: [],
    },
    yAxis: {
      type: 'value',
    },
    series: [
      {
        name: 'Income',
        type: 'line',
        smooth: true,
        data: [],
      },
    ],
  })

  async function fetchIncome () {
    const res = await fetch('http://localhost:4123/income')
    const data = await res.json()

    const branches = ['Branch A', 'Branch B', 'Branch C', 'Branch D']

    option.value.xAxis.data = data.income.map(i => i.date)

    option.value.series = branches.map(b => ({
      name: b,
      type: 'line',
      smooth: true,
      data: data.income.map(i => i.amounts[b]),
    }))
  }

  fetchIncome()
</script>

<style scoped>
.chart {
  height: 220px;
  width: 100%;
}
</style>
