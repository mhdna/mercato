<template>
  <div class="px-1 py-0" flat style="flex: 1 1 0; min-width: 0;">
    <v-card-title class="card-title pa-1 mb-n2">{{ title }}</v-card-title>
    <VChart autoresize :option="option" style="height: 140px; width: 100%; min-width: 0" />
  </div>
</template>

<script setup>
  import VChart from 'vue-echarts'

  defineProps({
    title: String,
    option: Object,
  })
</script>
