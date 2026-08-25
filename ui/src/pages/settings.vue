<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2 class="text-h6">Settings</h2>
    </div>

    <v-card max-width="480">
      <v-card-title class="text-subtitle-1">Branch activity</v-card-title>
      <v-card-subtitle>How live branch sales/returns/expenses are shown in the appbar</v-card-subtitle>
      <v-card-text>
        <div class="text-body-2 mb-2">Display mode</div>
        <v-btn-toggle
          v-model="displayMode"
          class="mb-4"
          color="primary"
          divided
          mandatory
        >
          <v-btn value="notification">Notification</v-btn>
          <v-btn value="appbar">Appbar messages</v-btn>
        </v-btn-toggle>

        <v-text-field
          v-model.number="messageSeconds"
          density="compact"
          hint="How long each message is shown before moving to the next one queued behind it"
          label="Message duration (seconds)"
          min="1"
          persistent-hint
          type="number"
        />
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { useSettingsStore } from '@/stores/settings'

  const settingsStore = useSettingsStore()

  const displayMode = computed({
    get: () => settingsStore.activityDisplayMode,
    set: value => settingsStore.setActivityDisplayMode(value),
  })

  const messageSeconds = computed({
    get: () => settingsStore.activityMessageSeconds,
    set: value => {
      const seconds = Number(value)
      if (Number.isFinite(seconds) && seconds > 0) {
        settingsStore.setActivityMessageSeconds(seconds)
      }
    },
  })
</script>
