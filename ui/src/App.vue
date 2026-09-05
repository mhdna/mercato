<template>
  <v-app>
    <UnsupportedDevice v-if="isNonDesktopBlocked" />
    <ErrorBoundary v-else>
      <router-view />
    </ErrorBoundary>
  </v-app>
</template>

<script lang="ts" setup>
  import { computed, provide, watch } from 'vue'
  import { THEME_KEY } from 'vue-echarts'
  import { useTheme } from 'vuetify'
  import ErrorBoundary from '@/components/ErrorBoundary.vue'
  import UnsupportedDevice from '@/components/UnsupportedDevice.vue'
  import { BLOCK_NON_DESKTOP } from '@/config'
  import { useAppStore } from '@/stores/app'

  const vuetifyTheme = useTheme()

  // App-wide ECharts theme: every <VChart> that doesn't set its own `theme`
  // prop inherits this and flips with Vuetify's light/dark mode. vue-echarts
  // reacts to the change and calls chart.setTheme(), so no re-mount needed.
  // Both names are registered in plugins/echarts.js.
  provide(THEME_KEY, computed(() => (vuetifyTheme.global.current.value.dark ? 'kashi-dark' : 'kashi-light')))
  const appStore = useAppStore()

  // Gate on the browser's user agent, not the viewport: a narrow desktop
  // window is still a desktop, and a phone held in landscape isn't. Covers
  // the usual mobile UA tokens plus iPadOS 13+, which reports a Mac UA and
  // is only distinguishable by having a touch screen.
  const isMobileOrTablet = (() => {
    const ua = navigator.userAgent || ''
    if (/Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|Tablet|Silk|Kindle|PlayBook/i.test(ua)) {
      return true
    }
    return /Macintosh/.test(ua) && navigator.maxTouchPoints > 1
  })()
  const isNonDesktopBlocked = computed(() => BLOCK_NON_DESKTOP && isMobileOrTablet)

  // Hand the preference straight to Vuetify. It accepts 'system' as a first
  // class value and follows the OS live on its own, so there's nothing else
  // to wire up here.
  watch(() => appStore.themePreference, pref => vuetifyTheme.change(pref), { immediate: true })
</script>
