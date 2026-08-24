<template>
  <v-card class="mx-auto pa-4" max-width="480" rounded="lg">
    <div class="d-flex align-center mb-2">
      <div>
        <div class="text-title-medium font-weight-bold">Page Views</div>
        <div class="text-body-small text-medium-emphasis">{{ periodLabel }}</div>
      </div>
      <v-spacer />
      <v-btn-toggle
        v-model="period"
        density="compact"
        mandatory
        rounded="lg"
        variant="outlined"
      >
        <v-btn value="weekly">Weekly</v-btn>
        <v-btn value="monthly">Monthly</v-btn>
        <v-btn value="quarterly">Quarterly</v-btn>
      </v-btn-toggle>
    </div>

    <v-sparkline
      animation
      auto-draw="once"
      auto-draw-duration="800"
      color="primary"
      line-width="2"
      :model-value="series[period]"
      smooth="4"
      stroke-linecap="round"
    />
  </v-card>
</template>
<script setup>
  import { computed, ref } from 'vue'

  const period = ref('monthly')

  const series = {
    weekly: [120, 340, 275, 410, 380, 295, 450],
    monthly: [640, 820, 550, 910, 770, 1050, 680, 1120, 860, 1240, 930, 1180],
    quarterly: [3100, 5800, 4200, 7600, 6100, 3800],
  }

  const periodLabel = computed(() => ({
    weekly: 'Last 7 days',
    monthly: 'Last 12 months',
    quarterly: 'Last 6 quarters',
  })[period.value])
</script>
