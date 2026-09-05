// Utilities
import { defineStore } from 'pinia'

// The user's theme choice: 'system' | 'light' | 'dark'. Vuetify understands
// 'system' natively (it tracks prefers-color-scheme itself), so this is the
// only value we persist -- App.vue hands it straight to theme.change().
function initialPreference () {
  const pref = localStorage.getItem('themePreference')
  if (pref === 'system' || pref === 'light' || pref === 'dark') {
    return pref
  }
  // Migrate the old key, which stored the already-resolved 'light'/'dark'.
  const legacy = localStorage.getItem('theme')
  if (legacy === 'light' || legacy === 'dark') {
    return legacy
  }
  return 'system'
}

export const useAppStore = defineStore('app', {
  state: () => ({
    themePreference: initialPreference(),
  }),

  actions: {
    setThemePreference (pref) {
      this.themePreference = pref
      localStorage.setItem('themePreference', pref)
      localStorage.removeItem('theme')
    },
  },
})
