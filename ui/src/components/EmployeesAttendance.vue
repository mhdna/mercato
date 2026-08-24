<template>
  <v-row class="h-100">
    <v-col>
      <v-sheet class="h-100">
        <v-calendar
          ref="calendar"
          color="primary"
          :events="events"
          :model-value="today"
          :now="today"
          type="week"
        />
      </v-sheet>
    </v-col>
  </v-row>
</template>

<script setup>
  import { onMounted, ref } from 'vue'

  const calendar = ref()

  const today = ref('2019-01-08')
  const events = [
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: '',
      start: '2019-01-07 09:00',
      end: '2019-01-07 10:00',
    },
    {
      name: ``,
      start: '2019-01-10 10:30',
      end: '2019-01-10 18:30',
    },
    {
      name: '',
      start: '2019-01-09 12:30',
      end: '2019-01-09 15:30',
    },
  ]

  onMounted(() => {
    calendar.value.scrollToTime('08:00')
  })
</script>
<style scoped>
.my-event {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 2px;
  background-color: #1867c0;
  color: #ffffff;
  border: 1px solid #1867c0;
  font-size: 12px;
  padding: 3px;
  cursor: pointer;
  margin-bottom: 1px;
  left: 4px;
  margin-right: 8px;
  position: relative;
}

.my-event.with-time {
  position: absolute;
  right: 4px;
  margin-right: 0px;
}
</style>
