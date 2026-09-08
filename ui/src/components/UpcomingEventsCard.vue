<template>
  <v-menu v-model="menu" :close-on-content-click="false" location="bottom center" :offset="8">
    <template #activator="{ props }">
      <v-card
        v-bind="props"
        class="d-flex align-center me-2"
        :class="mobile ? 'icon-pill' : 'py-1 px-2'"
        rounded="xl"
        style="cursor: pointer"
        variant="tonal"
      >
        <v-icon
          :class="{ 'me-2': !mobile }"
          :color="next ? next.color : undefined"
          :icon="next ? next.icon : 'mdi-calendar'"
          :size="mobile ? 20 : undefined"
        />
        <div v-if="!mobile" class="text-body-2 label">
          <template v-if="next">
            <span>{{ next.lead }}</span>
            <span class="ms-1">{{ next.name }}</span>
            <span class="ms-1">· {{ next.relative }}</span>
          </template>
          <span v-else>No events in {{ windowDays }} days</span>
        </div>
      </v-card>
    </template>

    <v-card max-width="400" min-width="300">
      <v-card-title class="text-subtitle-1 d-flex align-center">
        Upcoming events
        <v-spacer />
        <span class="text-caption text-medium-emphasis">next {{ windowDays }} days</span>
      </v-card-title>
      <v-divider />

      <v-list v-if="events.length > 0" density="compact" lines="two">
        <v-list-item
          v-for="event in events"
          :key="`${event.name}-${event.start.getTime()}`"
          :prepend-icon="event.icon"
          :subtitle="`${event.dateLabel} · ${event.relative}`"
          :title="event.name"
        >
          <template #prepend>
            <v-icon :color="event.color" :icon="event.icon" />
          </template>
        </v-list-item>
      </v-list>
      <v-card-text v-else class="text-medium-emphasis">
        Nothing on the retail calendar for the next {{ windowDays }} days.
      </v-card-text>
    </v-card>
  </v-menu>
</template>

<script setup>
  import { computed, onMounted, onUnmounted, ref } from 'vue'
  import { useDisplay } from 'vuetify'
  import { getEventsForRange } from '@/data/retailCalendarEvents'
  import { useSettingsStore } from '@/stores/settings'

  const settingsStore = useSettingsStore()
  const { mobile } = useDisplay()

  const menu = ref(false)
  // Ticks so the "in N days" labels stay honest across midnight without a
  // reload -- retail events are day-granular, so an hourly tick is plenty.
  const now = ref(Date.now())
  let clockTimer = null

  const windowDays = computed(() => settingsStore.upcomingEventsDays || 15)

  function isoDate (date) {
    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    return `${y}-${m}-${d}`
  }

  // Whole days from today (midnight) to the given date (midnight).
  function daysFromToday (date) {
    const today = new Date(now.value)
    today.setHours(0, 0, 0, 0)
    const target = new Date(date)
    target.setHours(0, 0, 0, 0)
    return Math.round((target - today) / 86_400_000)
  }

  function relativeLabel (start, end) {
    const startIn = daysFromToday(start)
    const endIn = daysFromToday(end)
    if (startIn <= 0 && endIn >= 0) return 'on now'
    if (startIn === 1) return 'tomorrow'
    if (startIn < 7) return `in ${startIn} days`
    const weeks = Math.round(startIn / 7)
    return `in ${weeks} week${weeks === 1 ? '' : 's'}`
  }

  function dateLabel (start, end) {
    const opts = { month: 'short', day: 'numeric' }
    const s = start.toLocaleDateString(undefined, opts)
    if (daysFromToday(start) === daysFromToday(end)) return s
    return `${s} – ${end.toLocaleDateString(undefined, opts)}`
  }

  // Events whose span overlaps [today, today + windowDays], soonest first.
  // getEventsForRange already expands the recurring retail calendar for us;
  // we just trim the past-dated tail and decorate with relative labels.
  const events = computed(() => {
    const start = new Date(now.value)
    const end = new Date(now.value)
    end.setDate(end.getDate() + windowDays.value)

    return getEventsForRange({ date: isoDate(start) }, { date: isoDate(end) })
      .filter(e => daysFromToday(e.end) >= 0 && daysFromToday(e.start) <= windowDays.value)
      .toSorted((a, b) => a.start - b.start)
      .map(e => ({
        ...e,
        relative: relativeLabel(e.start, e.end),
        dateLabel: dateLabel(e.start, e.end),
      }))
  })

  const next = computed(() => {
    const event = events.value[0]
    if (!event) return null
    return {
      ...event,
      lead: event.relative === 'on now' ? 'Now:' : 'Upcoming:',
    }
  })

  onMounted(() => {
    clockTimer = setInterval(() => {
      now.value = Date.now()
    }, 3_600_000)
  })

  onUnmounted(() => clearInterval(clockTimer))
</script>

<style scoped>
/* Matches SyncCard's mobile pill -- icon-only square. */
.icon-pill {
  width: 34px;
  height: 34px;
  justify-content: center;
}

.label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: min(32vw, 340px);
}
</style>
