<template>
  <v-row>
    <v-col>
      <v-sheet height="64">
        <v-toolbar flat>
          <v-btn
            class="me-4"
            color="grey-darken-2"
            variant="outlined"
            @click="setToday"
          >
            Today
          </v-btn>
          <v-btn
            color="grey-darken-2"
            icon
            size="small"
            variant="text"
            @click="prev"
          >
            <v-icon size="small">
              mdi-chevron-left
            </v-icon>
          </v-btn>
          <v-btn
            color="grey-darken-2"
            icon
            size="small"
            variant="text"
            @click="next"
          >
            <v-icon size="small">
              mdi-chevron-right
            </v-icon>
          </v-btn>
          <v-toolbar-title v-if="calendar">
            {{ calendar.title }}
          </v-toolbar-title>
          <v-menu location="bottom end">
            <template #activator="{ props }">
              <v-btn
                color="grey-darken-2"
                variant="outlined"
                v-bind="props"
              >
                <span>{{ typeToLabel[type] }}</span>
                <v-icon end>
                  mdi-menu-down
                </v-icon>
              </v-btn>
            </template>
            <v-list>
              <v-list-item @click="type = 'day'">
                <v-list-item-title>Day</v-list-item-title>
              </v-list-item>
              <v-list-item @click="type = 'week'">
                <v-list-item-title>Week</v-list-item-title>
              </v-list-item>
              <v-list-item @click="type = 'month'">
                <v-list-item-title>Month</v-list-item-title>
              </v-list-item>
              <v-list-item @click="type = '4day'">
                <v-list-item-title>4 days</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </v-toolbar>
      </v-sheet>
      <v-sheet height="835">
        <v-calendar
          ref="calendar"
          v-model="focus"
          color="primary"
          :event-color="getEventColor"
          :events="events"
          :type="type"
          @change="updateRange"
          @click:date="viewDay"
          @click:event="showEvent"
          @click:more="viewDay"
        />
        <v-menu
          v-model="selectedOpen"
          :activator="selectedElement"
          :close-on-content-click="false"
          location="end"
        >
          <v-card
            color="grey-lighten-4"
            flat
            min-width="350px"
          >
            <v-toolbar
              :color="selectedEvent.color"
              dark
            >
              <v-btn icon>
                <v-icon>mdi-pencil</v-icon>
              </v-btn>
              <v-toolbar-title v-html="selectedEvent.name" />
              <v-btn icon>
                <v-icon>mdi-heart</v-icon>
              </v-btn>
              <v-btn icon>
                <v-icon>mdi-dots-vertical</v-icon>
              </v-btn>
            </v-toolbar>
            <v-card-text>
              <span v-html="selectedEvent.details" />
            </v-card-text>
            <v-card-actions>
              <v-btn
                color="secondary"
                variant="text"
                @click="selectedOpen = false"
              >
                Cancel
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-menu>
      </v-sheet>
    </v-col>
  </v-row>
</template>
<script setup>
  import { onMounted, ref } from 'vue'

  const calendar = ref()

  const typeToLabel = {
    'month': 'Month',
    'week': 'Week',
    'day': 'Day',
    '4day': '4 Days',
  }
  const colors = ['blue', 'indigo', 'deep-purple', 'cyan', 'green', 'orange', 'grey darken-1']
  const names = ['Abbass', 'Sara', 'Fatima', 'Mahdi', 'Batoul', 'Hadi', 'Hussein', 'Mahdi']

  const focus = ref('')
  const type = ref('month')
  const selectedEvent = ref({})
  const selectedElement = ref(null)
  const selectedOpen = ref(false)
  const events = ref([])

  onMounted(() => {
    calendar.value.checkChange()
  })

  function viewDay (nativeEvent, { date }) {
    focus.value = date
    type.value = 'day'
  }

  function getEventColor (event) {
    return event.color
  }
  function setToday () {
    focus.value = ''
  }
  function prev () {
    calendar.value.prev()
  }
  function next () {
    calendar.value.next()
  }
  function showEvent (nativeEvent, { event }) {
    const open = () => {
      selectedEvent.value = event
      selectedElement.value = nativeEvent.target
      requestAnimationFrame(() => requestAnimationFrame(() => selectedOpen.value = true))
    }
    if (selectedOpen.value) {
      selectedOpen.value = false
      requestAnimationFrame(() => requestAnimationFrame(() => open()))
    } else {
      open()
    }
    nativeEvent.stopPropagation()
  }

  function updateRange ({ start, end }) {
    const _attendance = []
    const min = new Date(`${start.date}T00:00:00`)
    const max = new Date(`${end.date}T23:59:59`)
    const days = Math.floor((max.getTime() - min.getTime()) / 86_400_000) + 1

    for (let i = 0; i < days; i++) {
      const dayTimestamp = min.getTime() + (i * 86_400_000)
      const dayStart = new Date(dayTimestamp)

      for (let j = 0; j < 16; j++) {
        // Random start time between 10 AM and 11 AM
        const startHour = rnd(10, 11)
        const startMinute = rnd(0, 3) * 15 // 0, 15, 30, or 45
        const checkIn = new Date(dayStart)
        checkIn.setHours(startHour, startMinute, 0, 0)

        // 9 hours work + 1 hour break = 10 hours total
        const checkOut = new Date(checkIn.getTime() + (10 * 3_600_000))

        _attendance.push({
          name: names[rnd(0, names.length - 1)],
          start: checkIn,
          end: checkOut,
          color: colors[rnd(0, colors.length - 1)],
          timed: true,
        })
      }
    }
    events.value = _attendance
  }

  function rnd (a, b) {
    return Math.floor((b - a + 1) * Math.random()) + a
  }
</script>
