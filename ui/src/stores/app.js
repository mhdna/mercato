// Utilities
import { defineStore } from 'pinia'
import { useTheme } from 'vuetify'

export const useAppStore = defineStore('app', {
  state: () => ({
    theme: localStorage.getItem('theme') || (
      window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    ),
  }),

  actions: {
    toggleTheme () {
      const vuetifyTheme = useTheme()
      this.theme = this.theme === 'light' ? 'dark' : 'light'
      vuetifyTheme.global.name.value = this.theme
      localStorage.setItem('theme', this.theme)
    },

    setTheme (themeName) {
      const vuetifyTheme = useTheme()
      this.theme = themeName
      vuetifyTheme.global.name.value = themeName
      localStorage.setItem('theme', themeName)
    },
  },
})
