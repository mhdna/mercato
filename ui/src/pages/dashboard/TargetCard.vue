<template>
  <v-card class="target-card pa-3" flat rounded="lg">
    <div class="d-flex align-center justify-space-between">
      <span class="target-title font-weight-bold">{{ title }}</span>
      <span class="target-pill font-weight-medium">{{ timeLeft }}</span>
    </div>

    <v-divider class="target-divider my-2" />

    <div class="d-flex align-center justify-space-between">
      <div class="d-flex align-end">
        <span class="target-figure">{{ pct }}</span>
        <span class="target-figure-unit">%</span>
      </div>
      <div class="target-arrow d-flex align-center justify-center">
        <v-icon icon="mdi-arrow-top-right" size="13" />
      </div>
    </div>

    <div class="target-ticks mt-2">
      <span
        v-for="i in tickCount"
        :key="i"
        class="target-tick"
        :class="{ 'target-tick--dim': i > filledTicks }"
      />
    </div>
  </v-card>
</template>

<script setup>
  import { computed } from 'vue'

  const props = defineProps({
    title: { type: String, default: 'Productivity target' },
    timeLeft: { type: String, default: '2 weeks left' },
    pct: { type: Number, default: 78 },
  })

  const tickCount = 60
  const filledTicks = computed(() =>
    Math.round((Math.min(Math.max(props.pct, 0), 100) / 100) * tickCount),
  )
</script>

<style scoped>
  .target-card {
    background: #3949ab;
    color: #fff;
  }

  .target-title {
    font-size: 12px;
  }

  .target-pill {
    border: 1px solid rgba(255, 255, 255, 0.6);
    border-radius: 999px;
    padding: 1px 7px;
    font-size: 9px;
    white-space: nowrap;
  }

  .target-divider {
    border-color: rgba(255, 255, 255, 0.25);
    opacity: 1;
  }

  .target-figure {
    font-size: 24px;
    font-weight: 700;
    line-height: 1;
    font-variant-numeric: tabular-nums;
  }

  .target-figure-unit {
    font-size: 12px;
    font-weight: 500;
    line-height: 1;
    padding-bottom: 2px;
  }

  .target-arrow {
    width: 26px;
    height: 26px;
    border: 1.5px solid #fff;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .target-ticks {
    display: flex;
    align-items: stretch;
    gap: 2px;
  }

  .target-tick {
    flex: 1 1 0;
    height: 12px;
    background: #fff;
    border-radius: 1px;
  }

  .target-tick--dim {
    background: rgba(255, 255, 255, 0.28);
  }
</style>
