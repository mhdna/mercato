<template>
  <div class="page-root">
    <v-card class="pricing-card" flat>
      <v-card-title class="d-flex align-center ga-2 px-4 py-3">
        <v-icon icon="mdi-ticket-percent" />
        <span>Coupons</span>
      </v-card-title>
      <v-divider />
      <div class="pricing-content">
        <Coupons />
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import Coupons from '@/pages/coupons.vue'
</script>

<style scoped>
.page-root, .pricing-card { display: flex; flex: 1 1 auto; min-height: 0; flex-direction: column; }
.pricing-content { flex: 1 1 auto; min-width: 0; min-height: 0; overflow-y: auto; padding: 16px; }
</style>
