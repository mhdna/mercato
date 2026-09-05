<template>
  <div v-if="count > 0" class="bulk-bar d-flex flex-wrap align-center ga-2 px-3 py-2 mb-2 rounded">
    <v-icon icon="mdi-checkbox-multiple-marked-outline" size="20" />
    <span class="font-weight-medium">{{ count }} selected</span>
    <v-btn
      class="ms-1"
      density="comfortable"
      size="small"
      text="Clear"
      variant="text"
      @click="$emit('clear')"
    />
    <v-spacer />
    <v-alert
      v-if="error"
      class="py-1"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ error }}
    </v-alert>
    <v-btn
      color="error"
      :loading="loading"
      prepend-icon="mdi-delete"
      size="small"
      :text="`Delete ${count}`"
      variant="flat"
      @click="confirm = true"
    />

    <v-dialog v-model="confirm" max-width="420">
      <v-card>
        <v-card-title class="text-h6">Delete {{ count }} {{ count === 1 ? 'row' : 'rows' }}?</v-card-title>
        <v-card-text>
          The selected {{ count === 1 ? 'row' : 'rows' }} will be permanently deleted. This can't be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="confirm = false" />
          <v-btn
            color="error"
            :loading="loading"
            text="Delete"
            variant="flat"
            @click="onConfirm"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue'

  defineProps({
    count: { type: Number, default: 0 },
    loading: { type: Boolean, default: false },
    error: { type: String, default: '' },
  })
  const emit = defineEmits(['clear', 'confirm'])

  const confirm = ref(false)

  function onConfirm () {
    confirm.value = false
    emit('confirm')
  }
</script>

<style scoped>
.bulk-bar {
  background: rgba(var(--v-theme-primary), 0.06);
  border: 1px solid rgba(var(--v-theme-primary), 0.2);
}
</style>
