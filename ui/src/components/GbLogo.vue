<template>
  <div
    aria-label="GB"
    class="gb-logo"
    role="img"
    :style="{ '--gb-logo-image': `url(${gbImage})` }"
  />
</template>

<script setup lang="ts">
  import gbImage from '../../GB.png'
</script>

<style scoped>
  .gb-logo {
    aspect-ratio: 1;
    width: clamp(100px, 13vw, 180px);
    background-color: rgba(var(--v-theme-on-surface), 0.38);
    mask: var(--gb-logo-image) center / contain no-repeat;
    -webkit-mask: var(--gb-logo-image) center / contain no-repeat;
  }
</style>
