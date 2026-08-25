<template>
  <VChart autoresize class="chart" :option="option" />
</template>

<script setup>
  import VChart from 'vue-echarts'

  const option = ref({
    tooltip: {
      trigger: 'item',
      appendTo: 'body',
    },
    series: [
      {
        name: 'Access From',
        type: 'pie',
        radius: '50%',
        data: [
          { value: 1048, name: 'Search Engine' },
          { value: 735, name: 'Direct' },
          { value: 580, name: 'Email' },
          { value: 484, name: 'Union Ads' },
          { value: 300, name: 'Video Ads' },
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
        },
      },
    ],
  })
</script>

<style scoped>
.chart {
  height: 260px;
  width: 100%;
}
</style>
