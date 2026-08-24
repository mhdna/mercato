<template>
  <v-avatar size="150">
    <InitialsAvatar :initials="props.name[0]" />
    <!-- <v-img
      src="https://cdn.vuetifyjs.com/images/profiles/marcus.jpg"
      cover
    ></v-img> -->
  </v-avatar>
  <v-list-item :subtitle="title" :title="props.name" />
</template>

<script setup>
  const props = defineProps({
    name: String,
    title: String,
  })
</script>
