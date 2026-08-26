<template>
  <v-card class="w-100 h-100 d-flex flex-column">
    <v-toolbar color="transparent" density="compact" flat>
      <v-btn icon variant="text" @click="$refs.calendar.prev()">
        <v-icon icon="mdi-chevron-left" />
      </v-btn>
      <v-btn icon variant="text" @click="$refs.calendar.next()">
        <v-icon icon="mdi-chevron-right" />
      </v-btn>

      <v-toolbar-title class="text-subtitle-1 font-weight-medium">
        {{ title }}
      </v-toolbar-title>

      <v-btn prepend-icon="mdi-calendar-today" text="Today" variant="tonal" @click="goToToday" />
    </v-toolbar>

    <v-divider />

    <v-sheet class="flex-grow-1" style="min-height: 0;">
      <v-calendar
        ref="calendar"
        v-model="value"
        :event-color="getEventColor"
        :events="events"
        style="height: 100%;"
        :type="type"
        :weekdays="weekday"
        @change="getEvents"
      >
        <template #event="{ event }">
          <div class="d-flex align-center px-1" style="gap: 4px; overflow: hidden;">
            <v-icon v-if="event.icon" :icon="event.icon" size="14" />
            <span class="text-truncate">{{ event.name }}</span>
          </div>
        </template>
      </v-calendar>
    </v-sheet>
  </v-card>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { getEventsForRange } from '@/data/retailCalendarEvents'

  const type = ref('month')
  const weekday = ref([0, 1, 2, 3, 4, 5, 6])
  const value = ref('')
  const events = ref([])
  const rangeStart = ref(null)

  const title = computed(() => {
    if (!rangeStart.value) return ''
    return rangeStart.value.toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
  })

  function getEvents ({ start, end }) {
    rangeStart.value = new Date(`${start.date}T00:00:00`)
    events.value = getEventsForRange(start, end)
  }

  function getEventColor (event) {
    return event.color
  }

  function goToToday () {
    const today = new Date()
    const month = String(today.getMonth() + 1).padStart(2, '0')
    const day = String(today.getDate()).padStart(2, '0')
    value.value = `${today.getFullYear()}-${month}-${day}`
  }
</script>
