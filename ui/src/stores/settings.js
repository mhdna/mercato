// Utilities
import { defineStore } from 'pinia'

const DEFAULT_ACTIVITY_MESSAGE_SECONDS = 8
const DEFAULT_ACTIVITY_DISPLAY_MODE = 'notification'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    // How long a branch-activity message (sale/return/expense) is shown
    // before moving on to the next one queued behind it -- applies to both
    // display modes below. Per-viewer display preference, same
    // localStorage-backed pattern as theme in stores/app.js -- no backend
    // involved.
    activityMessageSeconds: Number(localStorage.getItem('activityMessageSeconds')) || DEFAULT_ACTIVITY_MESSAGE_SECONDS,
    // Where branch-activity messages are shown: 'notification' (queued
    // snackbar, bottom-right) or 'appbar' (cycles in place on SyncCard's
    // own line, no popups). Both SyncCard and BranchActivityToast read this
    // and only react to WS pushes when it's their mode.
    activityDisplayMode: localStorage.getItem('activityDisplayMode') || DEFAULT_ACTIVITY_DISPLAY_MODE,
  }),

  actions: {
    setActivityMessageSeconds (seconds) {
      this.activityMessageSeconds = seconds
      localStorage.setItem('activityMessageSeconds', String(seconds))
    },

    setActivityDisplayMode (mode) {
      this.activityDisplayMode = mode
      localStorage.setItem('activityDisplayMode', mode)
    },
  },
})
