<template>
  <div class="px-1 py-0 d-flex flex-column" flat style="flex: 1 1 0; min-width: 0;">
    <v-card-title class="card-title pa-1 mb-n2">{{ title }}</v-card-title>
    <VChart autoresize :option="option" :style="{ height, width: '100%', minWidth: 0 }" />
  </div>
</template>

<script setup>
  import VChart from 'vue-echarts'

  defineProps({
    title: String,
    option: Object,
    height: { type: String, default: '220px' },
  })
</script>
