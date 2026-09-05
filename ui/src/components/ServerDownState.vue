<template>
  <v-main class="server-down">
    <section class="server-down__content text-center">
      <v-icon class="server-down__icon" icon="mdi-server-network-off" size="72" />

      <h1 class="text-h4 font-weight-bold mt-6">Can't reach the server</h1>
      <p class="text-body-1 text-medium-emphasis mt-3 mb-0">
        {{ message }}
      </p>

      <v-btn
        class="mt-8"
        color="primary"
        :loading="retrying"
        prepend-icon="mdi-refresh"
        rounded="pill"
        size="large"
        @click="retry"
      >
        Try again
      </v-btn>
    </section>
  </v-main>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import { SERVER_DOWN_MESSAGE } from '@/composables/useApi'

  defineProps({
    message: {
      type: String,
      default: SERVER_DOWN_MESSAGE,
    },
  })

  // A full reload is the simplest reliable recovery: it re-runs auth, re-mounts
  // every page and re-issues the requests that failed while the server was off.
  const retrying = ref(false)
  function retry () {
    retrying.value = true
    window.location.reload()
  }
</script>

<style scoped>
.server-down {
  align-items: center;
  background:
    radial-gradient(circle at top, rgba(var(--v-theme-primary), 0.14), transparent 45%),
    rgb(var(--v-theme-background));
  display: flex;
  justify-content: center;
  min-height: 100dvh;
  padding: 24px;
}

.server-down__content {
  max-width: 380px;
  width: 100%;
}

.server-down__icon {
  color: rgba(var(--v-theme-on-surface), 0.35);
}
</style>
