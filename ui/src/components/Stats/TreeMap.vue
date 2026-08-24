<template>
  <VChart class="chart" :option="option" />
</template>

<script setup>
  import VChart from 'vue-echarts'

  const data = {
    children: [
      { name: 'option', value: 1 },
      { name: 'option.text', value: 1 },
      { name: 'option.text.style', value: 1 },
      { name: 'option.series', value: 1 },
      { name: 'option.series.type', value: 1 },
      { name: 'option.series.data', value: 1 },
    ],
  }

  const option = {
    tooltip: {},
    series: [
      {
        type: 'treemap',
        visibleMin: 300,
        data: data.children,
        levels: [
          {
            itemStyle: {
              borderColor: '#555',
              borderWidth: 4,
              gapWidth: 4,
            },
          },
          {
            colorSaturation: [0.3, 0.6],
            itemStyle: {
              borderColorSaturation: 0.7,
              gapWidth: 2,
              borderWidth: 2,
            },
          },
          {
            colorSaturation: [0.3, 0.5],
            itemStyle: {
              borderColorSaturation: 0.6,
              gapWidth: 1,
            },
          },
          {
            colorSaturation: [0.3, 0.5],
          },
        ],
      },
    ],
  }
</script>

<style scoped>
.chart {
  height: 400px;
  width: 100%;
}
</style>
