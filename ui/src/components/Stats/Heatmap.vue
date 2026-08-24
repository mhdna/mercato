<template>
  <VChart autoresize class="chart" :option="option" :theme="currentTheme" />
</template>

<script setup>
  import { computed, ref } from 'vue'
  import VChart from 'vue-echarts'
  import { useTheme } from 'vuetify'

  const theme = useTheme()
  const currentTheme = computed(() =>
    theme.global.current.value?.dark ? 'dark' : 'light',
  )

  function getVirtualData (year) {
    const date = +new Date(year + '-01-01')
    const end = +new Date(+year + 1 + '-01-01')
    const dayTime = 3600 * 24 * 1000
    const data = []
    for (let t = date; t < end; t += dayTime) {
      const yyyy = new Date(t).getFullYear()
      const mm = String(new Date(t).getMonth() + 1).padStart(2, '0')
      const dd = String(new Date(t).getDate()).padStart(2, '0')
      data.push([
        `${yyyy}-${mm}-${dd}`,
        Math.floor(Math.random() * 10_000),
      ])
    }
    return data
  }

  const option = ref({
    backgroundColor: 'transparent',
    tooltip: {},
    visualMap: {
      min: 0,
      max: 10_000,
      type: 'piecewise',
      orient: 'horizontal',
      left: 'center',
      bottom: 0,
    },
    calendar: {
      top: 30,
      left: 20,
      right: 20,
      cellSize: ['auto', 13],
      range: '2016',
      itemStyle: {
        borderWidth: 0.5,
      },
      yearLabel: { show: false },
    },
    series: {
      type: 'heatmap',
      coordinateSystem: 'calendar',
      data: getVirtualData('2016'),
    },
  })
</script>

<style scoped>
.chart {
  height: 300px;
}
</style>
