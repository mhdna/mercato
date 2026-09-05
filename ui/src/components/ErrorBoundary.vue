<template>
  <ServerDownState v-if="serverDown" />
  <slot v-else />
</template>

<script setup lang="ts">
  import { onErrorCaptured, ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import ServerDownState from '@/components/ServerDownState.vue'
  import { SERVER_DOWN_MESSAGE } from '@/composables/useApi'

  // A page's async setup that awaits authFetch throws all the way up here when
  // the server is unreachable. Without a boundary Vue tears the subtree down and
  // the user just sees a blank screen with console noise. Catch that one error
  // and render an explicit state instead; let everything else bubble to
  // app.config.errorHandler untouched.
  const serverDown = ref(false)

  onErrorCaptured(error => {
    if (error?.message === SERVER_DOWN_MESSAGE) {
      serverDown.value = true
      return false
    }
    return undefined
  })

  // If the user navigates (e.g. via the still-mounted nav drawer) after a
  // transient outage, drop the error state and let the target page try again.
  const route = useRoute()
  watch(() => route.fullPath, () => {
    serverDown.value = false
  })
</script>
