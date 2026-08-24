<script setup lang="ts">
  import WarehouseTreemap from '@/components/Stats/WarehouseTreemap.vue'
  import FilterBar from './FilterBar.vue'
</script>

<template>
  <!-- <Heatmap /> -->
  <v-container class="w-100 h-100 pa-0" fluid>
    <WarehouseTreemap />
  </v-container>
</template>
