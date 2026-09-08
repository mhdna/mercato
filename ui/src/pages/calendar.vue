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

      <v-spacer />

      <v-autocomplete
        v-model="selectedEvent"
        class="me-3"
        clearable
        density="compact"
        hide-details
        item-title="name"
        :items="searchEvents"
        placeholder="Search events"
        prepend-inner-icon="mdi-magnify"
        return-object
        variant="outlined"
        width="280"
        @update:model-value="goToEvent"
      />

      <v-btn
        class="me-2"
        prepend-icon="mdi-calendar-plus"
        text="Events"
        variant="tonal"
        @click="manageOpen = true"
      />
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

    <!-- Manage events: add / remove custom events and hide built-ins -->
    <v-dialog v-model="manageOpen" max-width="560" scrollable>
      <v-card>
        <v-card-title class="text-subtitle-1 d-flex align-center">
          Calendar events
          <v-spacer />
          <v-btn icon="mdi-close" size="small" variant="text" @click="manageOpen = false" />
        </v-card-title>
        <v-divider />

        <v-card-text>
          <v-alert
            v-if="eventsError"
            class="mb-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ eventsError }}
          </v-alert>

          <div class="text-subtitle-2 mb-2">Add an event</div>
          <v-row dense>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="form.name"
                density="compact"
                hide-details
                label="Name"
                variant="outlined"
              />
            </v-col>
            <v-col cols="6" sm="3">
              <v-text-field
                v-model="form.start_date"
                density="compact"
                hide-details
                label="Start"
                type="date"
                variant="outlined"
              />
            </v-col>
            <v-col cols="6" sm="3">
              <v-text-field
                v-model="form.end_date"
                density="compact"
                hide-details
                label="End"
                type="date"
                variant="outlined"
              />
            </v-col>
            <v-col cols="6" sm="6">
              <v-select
                v-model="form.color"
                density="compact"
                hide-details
                :items="colorOptions"
                label="Colour"
                variant="outlined"
              >
                <template #prepend-inner>
                  <v-icon :color="form.color" icon="mdi-circle" size="14" />
                </template>
              </v-select>
            </v-col>
            <v-col cols="6" sm="6">
              <v-text-field
                v-model="form.icon"
                density="compact"
                hide-details
                hint="Any mdi-* icon name"
                label="Icon"
                variant="outlined"
              >
                <template #prepend-inner>
                  <v-icon :icon="form.icon || 'mdi-calendar-star'" size="16" />
                </template>
              </v-text-field>
            </v-col>
          </v-row>
          <div class="d-flex justify-end mt-2">
            <v-btn
              :disabled="!canAdd"
              :loading="adding"
              prepend-icon="mdi-plus"
              size="small"
              text="Add event"
              variant="tonal"
              @click="addEvent"
            />
          </div>

          <template v-if="customEvents.length > 0">
            <v-divider class="my-4" />
            <div class="text-subtitle-2 mb-1">Your events</div>
            <v-list density="compact">
              <v-list-item
                v-for="ev in customEvents"
                :key="ev.id"
                :subtitle="dateRangeLabel(ev.start_date, ev.end_date)"
                :title="ev.name"
              >
                <template #prepend>
                  <v-icon :color="ev.color" :icon="ev.icon" />
                </template>
                <template #append>
                  <v-btn
                    color="error"
                    icon="mdi-delete"
                    size="small"
                    variant="text"
                    @click="removeEvent(ev.id)"
                  />
                </template>
              </v-list-item>
            </v-list>
          </template>

          <v-divider class="my-4" />
          <div class="text-subtitle-2 mb-1">Built-in events</div>
          <p class="text-caption text-medium-emphasis mb-1">
            Turn one off to hide it from the calendar and the appbar card.
          </p>
          <v-list density="compact">
            <v-list-item
              v-for="ev in builtinEvents"
              :key="ev.name"
              :title="ev.name"
            >
              <template #prepend>
                <v-icon :color="ev.color" :icon="ev.icon" />
              </template>
              <template #append>
                <v-switch
                  color="primary"
                  density="compact"
                  hide-details
                  :model-value="!hiddenSet.has(ev.name)"
                  @update:model-value="settingsStore.toggleBuiltinEventHidden(ev.name)"
                />
              </template>
            </v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup>
  import { computed, onMounted, ref, watch } from 'vue'
  import { useCalendarEvents } from '@/composables/useCalendarEvents'
  import { getEventsForRange, retailCalendarEvents } from '@/data/retailCalendarEvents'
  import { useSettingsStore } from '@/stores/settings'

  const settingsStore = useSettingsStore()
  const { listCalendarEvents, createCalendarEvent, deleteCalendarEvent } = useCalendarEvents()

  const type = ref('month')
  const weekday = ref([0, 1, 2, 3, 4, 5, 6])
  const value = ref('')
  const events = ref([])
  const rangeStart = ref(null)
  const selectedEvent = ref(null)

  // User-defined events + the currently visible calendar range, kept so we
  // can re-merge after an add / delete / hide without a page change.
  const customEvents = ref([])
  const lastRange = ref(null)

  const manageOpen = ref(false)
  const eventsError = ref('')
  const adding = ref(false)
  const form = ref(blankForm())

  const colorOptions = [
    'primary', 'red', 'pink', 'purple', 'indigo', 'blue',
    'teal', 'green', 'orange', 'brown', 'blue-grey',
  ]

  function blankForm () {
    return { name: '', icon: 'mdi-calendar-star', color: 'primary', start_date: '', end_date: '' }
  }

  const canAdd = computed(() => {
    const f = form.value
    return f.name.trim() !== '' && f.start_date !== '' && f.end_date !== '' && f.end_date >= f.start_date
  })

  const hiddenSet = computed(() => new Set(settingsStore.hiddenBuiltinEvents))

  const builtinEvents = computed(() =>
    retailCalendarEvents.map(e => ({ name: e.name, icon: e.icon, color: e.color })),
  )

  const searchEvents = computed(() => {
    const year = rangeStart.value?.getFullYear() ?? new Date().getFullYear()

    const builtin = retailCalendarEvents
      .filter(event => !hiddenSet.value.has(event.name))
      .flatMap(event => event.dates(year).map(({ start, end, allDay }) => ({
        name: event.name,
        start,
        end,
        allDay,
      })))

    const custom = customEvents.value.map(ev => ({
      name: ev.name,
      start: new Date(`${String(ev.start_date).slice(0, 10)}T00:00:00`),
      end: new Date(`${String(ev.end_date).slice(0, 10)}T00:00:00`),
      allDay: true,
    }))

    return [...builtin, ...custom]
  })

  const title = computed(() => {
    if (!rangeStart.value) return ''
    return rangeStart.value.toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
  })

  function renderRange () {
    if (!lastRange.value) return
    const { start, end } = lastRange.value
    events.value = getEventsForRange(start, end, {
      custom: customEvents.value,
      hidden: settingsStore.hiddenBuiltinEvents,
    })
  }

  function getEvents ({ start, end }) {
    rangeStart.value = new Date(`${start.date}T00:00:00`)
    lastRange.value = { start, end }
    renderRange()
  }

  function getEventColor (event) {
    return event.color
  }

  function dateRangeLabel (start, end) {
    const opts = { month: 'short', day: 'numeric', year: 'numeric' }
    const s = new Date(`${String(start).slice(0, 10)}T00:00:00`).toLocaleDateString(undefined, opts)
    const e = new Date(`${String(end).slice(0, 10)}T00:00:00`).toLocaleDateString(undefined, opts)
    return s === e ? s : `${s} – ${e}`
  }

  async function loadCustom () {
    try {
      customEvents.value = await listCalendarEvents()
      eventsError.value = ''
    } catch (error) {
      eventsError.value = error.message
    }
    renderRange()
  }

  async function addEvent () {
    if (!canAdd.value) return
    adding.value = true
    eventsError.value = ''
    try {
      await createCalendarEvent({
        name: form.value.name.trim(),
        icon: form.value.icon.trim() || 'mdi-calendar-star',
        color: form.value.color,
        start_date: form.value.start_date,
        end_date: form.value.end_date,
      })
      form.value = blankForm()
      await loadCustom()
    } catch (error) {
      eventsError.value = error.message
    } finally {
      adding.value = false
    }
  }

  async function removeEvent (id) {
    eventsError.value = ''
    try {
      await deleteCalendarEvent(id)
      await loadCustom()
    } catch (error) {
      eventsError.value = error.message
    }
  }

  // Re-merge whenever the hidden-built-ins set changes from the toggles.
  watch(() => settingsStore.hiddenBuiltinEvents, renderRange, { deep: true })

  onMounted(loadCustom)

  function goToToday () {
    const today = new Date()
    const month = String(today.getMonth() + 1).padStart(2, '0')
    const day = String(today.getDate()).padStart(2, '0')
    value.value = `${today.getFullYear()}-${month}-${day}`
  }

  function goToEvent (event) {
    if (!event?.start) return

    const date = event.start
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    value.value = `${date.getFullYear()}-${month}-${day}`
  }
</script>
