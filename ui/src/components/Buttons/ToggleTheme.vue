<template>
  <v-menu location="bottom end">
    <template #activator="{ props: menuProps }">
      <v-btn icon v-bind="menuProps">
        <v-icon :icon="activeOption.icon" />
      </v-btn>
    </template>

    <v-list density="compact" min-width="160">
      <v-list-item
        v-for="option in options"
        :key="option.value"
        :active="appStore.themePreference === option.value"
        :prepend-icon="option.icon"
        :title="option.title"
        @click="appStore.setThemePreference(option.value)"
      />
    </v-list>
  </v-menu>
</template>

<script lang="ts" setup>
  import { computed } from 'vue'
  import { useAppStore } from '@/stores/app'

  const appStore = useAppStore()

  const options = [
    { value: 'system', title: 'System', icon: 'mdi-theme-light-dark' },
    { value: 'light', title: 'Light', icon: 'mdi-white-balance-sunny' },
    { value: 'dark', title: 'Dark', icon: 'mdi-weather-night' },
  ]

  const activeOption = computed(
    () => options.find(o => o.value === appStore.themePreference) ?? options[0],
  )
</script>
