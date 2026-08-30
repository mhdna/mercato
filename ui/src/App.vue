<template>
  <v-app>
    <UnsupportedDevice v-if="isNonDesktopBlocked" />
    <router-view v-else />
  </v-app>
</template>

<script lang="ts" setup>
  import { computed, onMounted } from 'vue'
  import { useDisplay, useTheme } from 'vuetify'
  import UnsupportedDevice from '@/components/UnsupportedDevice.vue'
  import { BLOCK_NON_DESKTOP } from '@/config'
  import { useAppStore } from '@/stores/app'

  const vuetifyTheme = useTheme()
  const { lgAndUp } = useDisplay()
  const appStore = useAppStore()
  const isNonDesktopBlocked = computed(() => BLOCK_NON_DESKTOP && !lgAndUp.value)

  onMounted(() => {
    // Set initial theme from store
    vuetifyTheme.change(appStore.theme)

    // Listen for system theme changes and update if user hasn't set a preference
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleThemeChange = e => {
      const saved = localStorage.getItem('theme')
      if (!saved) {
        const newTheme = e.matches ? 'dark' : 'light'
        appStore.setTheme(newTheme)
        vuetifyTheme.change(newTheme)
      }
    }

    mediaQuery.addEventListener('change', handleThemeChange)
  })
</script>
