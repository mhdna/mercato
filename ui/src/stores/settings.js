// Utilities
import { defineStore } from 'pinia'

const DEFAULT_ACTIVITY_MESSAGE_SECONDS = 8

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    // How long the appbar SyncCard shows each branch-activity message
    // (sale/return/expense) before cycling to the next one queued behind
    // it. Per-viewer display preference, same localStorage-backed pattern
    // as theme in stores/app.js -- no backend involved.
    activityMessageSeconds: Number(localStorage.getItem('activityMessageSeconds')) || DEFAULT_ACTIVITY_MESSAGE_SECONDS,
  }),

  actions: {
    setActivityMessageSeconds (seconds) {
      this.activityMessageSeconds = seconds
      localStorage.setItem('activityMessageSeconds', String(seconds))
    },
  },
})
