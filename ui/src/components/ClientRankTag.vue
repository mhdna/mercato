<template>
  <!-- TODO: color based on rank -->
  <div class="d-flex mt-2 mx-3 px-2" style="background-color: gold; height: 25px; width: 120px;">
    <v-icon class="me-2" icon="mdi-star" />
    <div>
      {{ props.rank }}
    </div>
  </div>
</template>

<script setup>
  const props = defineProps({
    rank: {
      type: String,
      required: true,
    },
  })
</script>
