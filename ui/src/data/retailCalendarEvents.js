function nthWeekdayOfMonth (year, month, weekday, n) {
  const first = new Date(year, month, 1)
  const day = 1 + ((weekday - first.getDay() + 7) % 7) + (n - 1) * 7
  return new Date(year, month, day)
}

function addDays (date, days) {
  const d = new Date(date)
  d.setDate(d.getDate() + days)
  return d
}

// Mean length of a tabular Hijri year, used only to approximate where a Hijri
// holiday falls in the Gregorian calendar. The real date depends on moon
// sighting and can differ by a day or two - events below are labelled
// "(approx.)" for that reason.
const HIJRI_YEAR_DAYS = 354.367_07
const EID_AL_FITR_ANCHOR = new Date(2024, 3, 10) // 1 Shawwal 1445 AH
const EID_AL_ADHA_ANCHOR = new Date(2024, 5, 16) // 10 Dhu al-Hijjah 1445 AH

function approximateHijriHoliday (anchor, year) {
  const roughSteps = Math.round((year - anchor.getFullYear()) * 365.2425 / HIJRI_YEAR_DAYS)
  for (const delta of [0, 1, -1, 2, -2]) {
    const candidate = addDays(anchor, Math.round((roughSteps + delta) * HIJRI_YEAR_DAYS))
    if (candidate.getFullYear() === year) {
      return candidate
    }
  }
  return null
}

export const retailCalendarEvents = [
  {
    name: 'New Year Sale',
    color: 'indigo',
    icon: 'mdi-firework',
    dates: year => [{ start: new Date(year, 0, 1), end: new Date(year, 0, 7), allDay: true }],
  },
  {
    name: 'Valentine’s Day',
    color: 'pink',
    icon: 'mdi-heart',
    dates: year => [{ start: new Date(year, 1, 14), end: new Date(year, 1, 14), allDay: true }],
  },
  {
    name: 'Mother’s Day',
    color: 'deep-purple',
    icon: 'mdi-flower-tulip',
    dates: year => {
      const d = nthWeekdayOfMonth(year, 4, 0, 2) // 2nd Sunday of May
      return [{ start: d, end: d, allDay: true }]
    },
  },
  {
    name: 'Summer Sale',
    color: 'orange',
    icon: 'mdi-white-balance-sunny',
    dates: year => [{ start: new Date(year, 5, 1), end: new Date(year, 5, 30), allDay: true }],
  },
  {
    name: 'Back to School',
    color: 'teal',
    icon: 'mdi-school',
    dates: year => [{ start: new Date(year, 7, 15), end: new Date(year, 8, 5), allDay: true }],
  },
  {
    name: 'Halloween',
    color: 'deep-orange',
    icon: 'mdi-halloween',
    dates: year => [{ start: new Date(year, 9, 31), end: new Date(year, 9, 31), allDay: true }],
  },
  {
    name: 'Black Friday',
    color: 'grey-darken-3',
    icon: 'mdi-sale',
    dates: year => {
      const thanksgiving = nthWeekdayOfMonth(year, 10, 4, 4) // 4th Thursday of Nov
      const d = addDays(thanksgiving, 1)
      return [{ start: d, end: d, allDay: true }]
    },
  },
  {
    name: 'Cyber Monday',
    color: 'blue-darken-2',
    icon: 'mdi-laptop',
    dates: year => {
      const thanksgiving = nthWeekdayOfMonth(year, 10, 4, 4)
      const d = addDays(thanksgiving, 4)
      return [{ start: d, end: d, allDay: true }]
    },
  },
  {
    name: 'Christmas Day',
    color: 'green-darken-2',
    icon: 'mdi-pine-tree',
    dates: year => [{ start: new Date(year, 11, 25), end: new Date(year, 11, 25), allDay: true }],
  },
  {
    name: 'Winter Clearance Sale',
    color: 'cyan-darken-1',
    icon: 'mdi-snowflake',
    dates: year => [{ start: new Date(year, 11, 26), end: new Date(year + 1, 0, 10), allDay: true }],
  },
  {
    name: 'Eid al-Fitr (approx.)',
    color: 'amber-darken-2',
    icon: 'mdi-moon-waning-crescent',
    dates: year => {
      const d = approximateHijriHoliday(EID_AL_FITR_ANCHOR, year)
      return d ? [{ start: d, end: addDays(d, 2), allDay: true }] : []
    },
  },
  {
    name: 'Eid al-Adha (approx.)',
    color: 'brown-darken-1',
    icon: 'mdi-moon-waning-crescent',
    dates: year => {
      const d = approximateHijriHoliday(EID_AL_ADHA_ANCHOR, year)
      return d ? [{ start: d, end: addDays(d, 3), allDay: true }] : []
    },
  },
]

// `custom` is the rows from useCalendarEvents().listCalendarEvents() --
// user-defined one-off events, each { id, name, icon, color, start_date,
// end_date } with ISO date strings. `hidden` is a list of built-in event
// names the user has chosen to hide.
export function getEventsForRange (start, end, { custom = [], hidden = [] } = {}) {
  const startDate = new Date(`${start.date}T00:00:00`)
  const endDate = new Date(`${end.date}T23:59:59`)
  const hiddenSet = new Set(hidden)
  const events = []

  for (let year = startDate.getFullYear() - 1; year <= endDate.getFullYear() + 1; year++) {
    for (const event of retailCalendarEvents) {
      if (hiddenSet.has(event.name)) {continue}
      for (const { start: s, end: e, allDay } of event.dates(year)) {
        if (e >= startDate && s <= endDate) {
          events.push({
            name: event.name,
            start: s,
            end: e,
            color: event.color,
            icon: event.icon,
            timed: !allDay,
            builtin: true,
          })
        }
      }
    }
  }

  for (const row of custom) {
    const s = new Date(`${String(row.start_date).slice(0, 10)}T00:00:00`)
    const e = new Date(`${String(row.end_date).slice(0, 10)}T23:59:59`)
    if (Number.isNaN(s.getTime()) || Number.isNaN(e.getTime())) {continue}
    if (e >= startDate && s <= endDate) {
      events.push({
        id: row.id,
        name: row.name,
        start: s,
        end: e,
        color: row.color || 'primary',
        icon: row.icon || 'mdi-calendar-star',
        timed: false,
        custom: true,
      })
    }
  }

  return events
}
