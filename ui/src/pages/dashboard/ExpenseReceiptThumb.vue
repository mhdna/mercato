<template>
  <span v-if="!imageId" class="text-medium-emphasis">—</span>

  <v-menu
    v-else
    location="start"
    :open-delay="80"
    open-on-hover
  >
    <template #activator="{ props: activator }">
      <v-avatar
        v-bind="activator"
        rounded="sm"
        size="36"
      >
        <v-img v-if="url" cover :src="url" />
        <v-icon v-else icon="mdi-image" size="18" />
      </v-avatar>
    </template>

    <v-card v-if="url" flat>
      <v-img
        max-height="360"
        max-width="360"
        :src="url"
        width="360"
      />
    </v-card>
  </v-menu>
</template>

<script setup>
  import { onUnmounted, ref, watch } from 'vue'
  import { useExpenseImages } from '@/composables/useExpenseImages'

  const props = defineProps({
    // branch_expense_images.id, or 0/null when the expense has no receipt.
    imageId: {
      type: Number,
      default: null,
    },
  })

  const { fetchExpenseImageObjectUrl } = useExpenseImages()
  const url = ref('')

  function revoke () {
    if (url.value) {
      URL.revokeObjectURL(url.value)
      url.value = ''
    }
  }

  watch(() => props.imageId, async id => {
    revoke()
    if (!id) return
    try {
      url.value = await fetchExpenseImageObjectUrl(id)
    } catch {
      // Leave unset -- the avatar falls back to a placeholder icon.
    }
  }, { immediate: true })

  onUnmounted(revoke)
</script>
