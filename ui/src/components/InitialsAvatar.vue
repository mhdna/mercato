<template>
  <v-avatar class="text-white" color="blue-darken-4" size="180">
    <span class="text-h2"> {{ initials }} Z </span>
  </v-avatar>
</template>

<script setup>
  import { defineProps } from 'vue'

  defineProps({
    initials: {
      type: String,
      required: true,
    },
  })
</script>
