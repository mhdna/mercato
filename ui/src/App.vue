<template>
  <v-app>
    <router-view />
  </v-app>
</template>

<script setup>
  import { onMounted } from 'vue'
  import { useTheme } from 'vuetify'
  import { useAppStore } from '@/stores/app'

  const vuetifyTheme = useTheme()
  const appStore = useAppStore()

  onMounted(() => {
    // Set initial theme from store
    vuetifyTheme.global.name.value = appStore.theme

    // Listen for system theme changes and update if user hasn't set a preference
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleThemeChange = e => {
      const saved = localStorage.getItem('theme')
      if (!saved) {
        const newTheme = e.matches ? 'dark' : 'light'
        appStore.setTheme(newTheme)
      }
    }

    mediaQuery.addEventListener('change', handleThemeChange)
  })
</script>
