<template>
  <div class="detail-list">
    <template v-for="([label, value]) in rows" :key="label">
      <div class="text-caption text-medium-emphasis">{{ label }}</div>
      <div class="text-body-2 mb-2 text-truncate">{{ value == null || value === '' ? '—' : value }}</div>
    </template>
  </div>
</template>

<script setup>
  defineProps({
    // [[label, value], ...]
    rows: {
      type: Array,
      default: () => [],
    },
  })
</script>
