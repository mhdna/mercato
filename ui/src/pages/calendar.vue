<template>
  <div class="w-100 h-100">
    <v-sheet class="d-flex" tile>
      <v-spacer />
      <v-btn class="ma-2" icon variant="text" @click="$refs.calendar.prev()">
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>
      <v-btn class="ma-2" icon variant="text" @click="$refs.calendar.next()">
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
      <v-spacer />
    </v-sheet>
    <v-sheet height="600">
      <v-calendar
        ref="calendar"
        v-model="value"
        :event-color="getEventColor"
        :event-overlap-mode="mode"
        :event-overlap-threshold="30"
        :events="events"
        :type="type"
        :weekdays="weekday"
        @change="getEvents"
      />
    </v-sheet>
  </div>
</template>
<script setup>
  import { ref } from 'vue'

  const type = ref('month')
  const weekday = ref([0, 1, 2, 3, 4, 5, 6])
  const value = ref('')
  const events = ref([])
  const colors = ['blue', 'indigo', 'deep-purple', 'cyan', 'green', 'orange', 'grey-darken-1']
  const names = ['Meeting', 'Holiday', 'PTO', 'Travel', 'Event', 'Birthday', 'Conference', 'Party']

  function rnd (a, b) {
    return Math.floor((b - a + 1) * Math.random()) + a
  }

  function getEvents ({ start, end }) {
    const evts = []

    const min = new Date(`${start.date}T00:00:00`)
    const max = new Date(`${end.date}T23:59:59`)
    const days = (max.getTime() - min.getTime()) / 86_400_000
    const eventCount = rnd(days, days + 20)

    for (let i = 0; i < eventCount; i++) {
      const allDay = rnd(0, 3) === 0
      const firstTimestamp = rnd(min.getTime(), max.getTime())
      const first = new Date(firstTimestamp - (firstTimestamp % 900_000))
      const secondTimestamp = rnd(2, allDay ? 288 : 8) * 900_000
      const second = new Date(first.getTime() + secondTimestamp)

      evts.push({
        name: names[rnd(0, names.length - 1)],
        start: first,
        end: second,
        color: colors[rnd(0, colors.length - 1)],
        timed: !allDay,
      })
    }

    events.value = evts
  }

  function getEventColor (event) {
    return event.color
  }
</script>
