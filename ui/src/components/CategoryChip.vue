<template>
  <span v-if="category" class="d-inline-flex align-center ga-2">
    <v-avatar :color="color" rounded="md" size="26" variant="tonal">
      <v-icon :color="color" :icon="icon" size="16" />
    </v-avatar>
    <span>{{ category.name }}</span>
  </span>
  <span v-else class="text-medium-emphasis">—</span>
</template>

<script setup>
  import { computed } from 'vue'
  import { categoryColor, categoryIcon } from '@/data/expenseCategoryIcons'

  const props = defineProps({
    category: { type: Object, default: null },
  })

  const icon = computed(() => categoryIcon(props.category))
  const color = computed(() => categoryColor(props.category))
</script>
