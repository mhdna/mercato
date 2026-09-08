// Utilities
import { defineStore } from 'pinia'
import { useAppSettings } from '@/composables/useAppSettings'

const DEFAULT_ACTIVITY_MESSAGE_SECONDS = 8
const DEFAULT_ACTIVITY_DISPLAY_MODE = 'notification'
const DEFAULT_FINANCIALS_HIGH_SEASON_MONTHS = [2, 5, 6, 9, 11, 12]
const DEFAULT_BARCODE_LABEL_WIDTH = 288
const DEFAULT_BARCODE_LABEL_HEIGHT = 144
const DEFAULT_UPCOMING_EVENTS_DAYS = 15
const DEFAULT_UPCOMING_EVENTS_MENU_DAYS = 180

// All app settings live in the single global app_settings DB row (see
// db/migrations/000050_create_app_settings + 000051_add_activity_settings)
// -- nothing here is localStorage-backed. State starts at sane defaults so
// the appbar (BranchActivityToast/SyncCard) has something to render before
// `init()` resolves; `init()` is called once from the authenticated layout
// (layouts/default.vue) and every setter below persists immediately.
export const useSettingsStore = defineStore('settings', {
  state: () => ({
    activityMessageSeconds: DEFAULT_ACTIVITY_MESSAGE_SECONDS,
    activityDisplayMode: DEFAULT_ACTIVITY_DISPLAY_MODE,
    financialsHighSeasonMonths: DEFAULT_FINANCIALS_HIGH_SEASON_MONTHS,
    barcodeLabelWidth: DEFAULT_BARCODE_LABEL_WIDTH,
    barcodeLabelHeight: DEFAULT_BARCODE_LABEL_HEIGHT,
    upcomingEventsDays: DEFAULT_UPCOMING_EVENTS_DAYS,
    upcomingEventsMenuDays: DEFAULT_UPCOMING_EVENTS_MENU_DAYS,
    hiddenBuiltinEvents: [],
    loaded: false,
  }),

  actions: {
    async init () {
      if (this.loaded) {
        return
      }
      const { getAppSettings } = useAppSettings()
      const settings = await getAppSettings().catch(() => null)
      if (settings) {
        this.activityMessageSeconds = settings.activity_message_seconds
        this.activityDisplayMode = settings.activity_display_mode
        this.financialsHighSeasonMonths = settings.financials_high_season_months
        this.barcodeLabelWidth = settings.barcode_label_width ?? DEFAULT_BARCODE_LABEL_WIDTH
        this.barcodeLabelHeight = settings.barcode_label_height ?? DEFAULT_BARCODE_LABEL_HEIGHT
        this.upcomingEventsDays = settings.upcoming_events_days ?? DEFAULT_UPCOMING_EVENTS_DAYS
        this.upcomingEventsMenuDays = settings.upcoming_events_menu_days ?? DEFAULT_UPCOMING_EVENTS_MENU_DAYS
        this.hiddenBuiltinEvents = settings.hidden_builtin_events ?? []
      }
      this.loaded = true
    },

    async persist () {
      const { updateAppSettings } = useAppSettings()
      await updateAppSettings({
        financials_high_season_months: this.financialsHighSeasonMonths,
        activity_message_seconds: this.activityMessageSeconds,
        activity_display_mode: this.activityDisplayMode,
        barcode_label_width: this.barcodeLabelWidth,
        barcode_label_height: this.barcodeLabelHeight,
        upcoming_events_days: this.upcomingEventsDays,
        upcoming_events_menu_days: this.upcomingEventsMenuDays,
        hidden_builtin_events: this.hiddenBuiltinEvents,
      })
    },

    async setUpcomingEventsDays (days) {
      this.upcomingEventsDays = days
      await this.persist()
    },

    async setUpcomingEventsMenuDays (days) {
      this.upcomingEventsMenuDays = days
      await this.persist()
    },

    async setHiddenBuiltinEvents (names) {
      this.hiddenBuiltinEvents = names
      await this.persist()
    },

    async toggleBuiltinEventHidden (name) {
      const hidden = new Set(this.hiddenBuiltinEvents)
      if (hidden.has(name)) {
        hidden.delete(name)
      } else {
        hidden.add(name)
      }
      this.hiddenBuiltinEvents = [...hidden]
      await this.persist()
    },

    async setBarcodeLabelSize (width, height) {
      this.barcodeLabelWidth = width
      this.barcodeLabelHeight = height
      await this.persist()
    },

    async setActivityMessageSeconds (seconds) {
      this.activityMessageSeconds = seconds
      await this.persist()
    },

    async setActivityDisplayMode (mode) {
      this.activityDisplayMode = mode
      await this.persist()
    },

    async setFinancialsHighSeasonMonths (months) {
      this.financialsHighSeasonMonths = months
      await this.persist()
    },
  },
})
