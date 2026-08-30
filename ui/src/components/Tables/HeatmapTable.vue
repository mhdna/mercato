<template>
  <div ref="wrapRef" class="heatmap-wrap">
    <table class="heatmap-table" @mouseleave="hoveredColumn = null">
      <thead>
        <tr>
          <th class="sticky-col day-col year-nav">
            <v-btn
              aria-label="Previous year"
              density="compact"
              icon="mdi-chevron-left"
              size="x-small"
              variant="text"
              @click="selectedYear--"
            />
            <span>{{ selectedYear }}</span>
            <v-btn
              aria-label="Next year"
              density="compact"
              :disabled="selectedYear >= currentYear"
              icon="mdi-chevron-right"
              size="x-small"
              variant="text"
              @click="selectedYear++"
            />
          </th>
          <th
            v-for="(month, mi) in months"
            :key="month"
            :class="{ 'column-hover': hoveredColumn === mi }"
            @mouseenter="hoveredColumn = mi"
          >
            {{ month }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in 31" :key="d">
          <td class="sticky-col day-col">{{ d }}</td>
          <td
            v-for="(month, mi) in months"
            :key="month"
            :ref="el => setCellRef(mi, d, el)"
            :class="{
              'cell-flash': isHighlighted(mi, d),
              'cell-today': isToday(mi, d),
              'column-hover': hoveredColumn === mi,
            }"
            :style="d <= daysInMonth[mi] ? cellStyle(get(mi, d)) : {}"
            @mouseenter="hoveredColumn = mi"
          >
            {{ d <= daysInMonth[mi] ? format(get(mi, d)) : "" }}
          </td>
        </tr>
      </tbody>
      <tfoot>
        <tr>
          <td class="sticky-col day-col sticky-foot">Total</td>
          <td
            v-for="(month, mi) in months"
            :key="month"
            :class="['sticky-foot', { 'column-hover': hoveredColumn === mi }]"
            @mouseenter="hoveredColumn = mi"
          >
            {{ format(monthTotals[mi]) }}
          </td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<script setup>
  import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { useDailyIncome } from '@/composables/useDailyIncome'
  import { useIncomeFilters } from '@/pages/dashboard/composables/useIncomeFilters'

  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ]
  const currentYear = new Date().getFullYear()
  const selectedYear = ref(currentYear)
  const daysInMonth = computed(() => {
    const year = selectedYear.value
    const isLeapYear = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0)
    return [31, isLeapYear ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
  })

  const { getDailyIncome } = useDailyIncome()
  const { selectedBranchId } = useIncomeFilters()
  const { ensureConnected, onMessage } = useAdminSocket()

  // data[monthIndex][day] = number, cents from the API converted to whole
  // currency units for display, same shape the heatmap always rendered.
  const data = ref(months.map((_, mi) => Array.from({ length: daysInMonth.value[mi] }, () => 0)))

  // Cells whose value changed on the most recent load, pulsed for 2s so a
  // synced sale is visible without staring at the grid.
  const highlighted = ref(new Set())
  const highlightTimers = new Map()
  const hoveredColumn = ref(null)

  function flashCell (mi, day) {
    const key = `${mi}-${day}`
    highlighted.value.add(key)
    highlighted.value = new Set(highlighted.value)
    clearTimeout(highlightTimers.get(key))
    highlightTimers.set(key, setTimeout(() => {
      highlighted.value.delete(key)
      highlighted.value = new Set(highlighted.value)
      highlightTimers.delete(key)
    }, 2000))
  }

  const isHighlighted = (mi, day) => highlighted.value.has(`${mi}-${day}`)

  // Scroll the grid so today's cell is in view on first load -- the sticky
  // day column/month header stay fixed, so the offset is subtracted out
  // rather than relying on scrollIntoView (which doesn't know about them).
  const wrapRef = ref(null)
  const today = new Date()
  const todayMonthIndex = today.getMonth()
  const todayDay = today.getDate()
  let todayCellEl = null
  let scrolledToToday = false

  function setCellRef (mi, day, el) {
    if (mi === todayMonthIndex && day === todayDay) todayCellEl = el
  }

  const isToday = (mi, day) => selectedYear.value === currentYear && mi === todayMonthIndex && day === todayDay

  async function scrollToToday () {
    if (scrolledToToday) return
    await nextTick()
    if (!wrapRef.value || !todayCellEl) return
    scrolledToToday = true
    const wrap = wrapRef.value
    const stickyColWidth = wrap.querySelector('.day-col')?.offsetWidth ?? 0
    const stickyHeaderHeight = wrap.querySelector('thead')?.offsetHeight ?? 0
    wrap.scrollLeft = todayCellEl.offsetLeft - stickyColWidth - (wrap.clientWidth - stickyColWidth) / 2 + todayCellEl.offsetWidth / 2
    wrap.scrollTop = todayCellEl.offsetTop - stickyHeaderHeight - (wrap.clientHeight - stickyHeaderHeight) / 2 + todayCellEl.offsetHeight / 2
  }

  // highlightChanges is false for a branch-filter switch (a whole new grid,
  // nothing to flash) and true for a live socket-triggered refresh.
  async function load (highlightChanges = false) {
    const year = selectedYear.value
    const days = await getDailyIncome(year, selectedBranchId.value).catch(() => [])
    if (year !== selectedYear.value) return
    const next = months.map((_, mi) => Array.from({ length: daysInMonth.value[mi] }, () => 0))
    for (const entry of days) {
      const d = new Date(entry.day)
      next[d.getUTCMonth()][d.getUTCDate() - 1] = entry.total / 100
    }
    const previous = data.value
    data.value = next
    if (highlightChanges) {
      for (const [mi, month] of next.entries()) {
        for (const [day, value] of month.entries()) {
          if (value !== previous[mi]?.[day]) flashCell(mi, day + 1)
        }
      }
    }
    scrollToToday()
  }

  watch([selectedBranchId, selectedYear], () => load(false), { immediate: true })

  // Branch sync pushes new invoices over the admin socket -- without this,
  // today's cell only updates on the next full page load. SyncCard listens
  // to the same event for its own refresh; see useAdminSocket.js.
  let unsubscribe = null
  onMounted(() => {
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (message.type === 'branch_invoice_created') {
        load(true)
      }
    })
  })
  onUnmounted(() => {
    unsubscribe?.()
    for (const timer of highlightTimers.values()) clearTimeout(timer)
  })

  const get = (mi, day) => data.value[mi][day - 1] ?? 0

  const monthTotals = computed(() => data.value.map(month => month.reduce((sum, v) => sum + v, 0)))

  const allValues = computed(() => data.value.flat())
  const min = computed(() => Math.min(...allValues.value))
  const max = computed(() => Math.max(...allValues.value))

  const format = v => '$' + v.toLocaleString()

  function cellStyle (v) {
    const t = (v - min.value) / (max.value - min.value || 1)
    const alpha = 0.08 + t * 0.85
    return {
      backgroundColor: `rgba(var(--v-theme-primary), ${alpha})`,
      color: t > 0.55 ? '#fff' : 'inherit',
    }
  }
</script>

<style scoped>
.heatmap-wrap {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  height: 100%;
  overflow: auto;
  border: 1px solid rgba(0, 0, 0, 0.12);
}

.heatmap-table {
  border-collapse: collapse;
  width: 100%;
  font-size: 13px;
}

.heatmap-table th,
.heatmap-table td {
  padding: 6px 12px;
  text-align: right;
  white-space: nowrap;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}

.heatmap-table .column-hover {
  background-image: linear-gradient(
    rgba(var(--v-theme-primary), 0.18),
    rgba(var(--v-theme-primary), 0.18)
  );
}

.heatmap-table thead th {
  position: sticky;
  top: 0;
  text-align: center;
  font-weight: 500;
  background: rgb(var(--v-theme-surface));
  z-index: 2;
}

.sticky-col {
  position: sticky;
  left: 0;
  text-align: left;
  background: rgb(var(--v-theme-surface));
  z-index: 1;
}

.heatmap-table thead .sticky-col {
  z-index: 3;
}

.day-col {
  width: 126px;
  min-width: 126px;
  max-width: 126px;
  font-weight: 500;
}

.year-nav {
  padding: 2px 4px !important;
  text-align: center !important;
}

.year-nav span {
  display: inline-block;
  min-width: 38px;
  font-variant-numeric: tabular-nums;
}

.heatmap-table tfoot td {
  font-weight: 600;
  border-bottom: none;
  border-top: 2px solid rgba(0, 0, 0, 0.12);
}

.sticky-foot {
  position: sticky;
  bottom: 0;
  background: #ffc107;
  color: #212121;
  z-index: 2;
}

.heatmap-table tfoot .sticky-col.sticky-foot {
  z-index: 3;
}

.cell-today {
  outline: 3px solid #f44336;
  outline-offset: -3px;
}

/* The flash is a separate overlay layer (::before), not an animation of the
   cell's own background-color. Animating background-color straight to
   "transparent" removes the cell's paint entirely for that instant, so what
   shows through is the *table's* background (white/dark surface) rather
   than the heat color underneath -- a visible flash-to-white before the
   heat color snaps back in. Layering the yellow on top and fading its
   opacity instead lets the heat color stay put underneath the whole time. */
.cell-flash {
  position: relative;
  z-index: 1;
  animation: cell-pop 2s cubic-bezier(0.22, 1, 0.36, 1);
}

.cell-flash::before {
  content: "";
  position: absolute;
  inset: 0;
  z-index: -1;
  background-color: #ffc107;
  animation: cell-flash-overlay 2s cubic-bezier(0.22, 1, 0.36, 1);
}

@keyframes cell-pop {
  0% {
    color: #212121;
    box-shadow: 0 0 0 0 rgba(255, 193, 7, 0.7);
    transform: scale(1.12);
  }
  30% {
    color: #212121;
    box-shadow: 0 0 10px 3px rgba(255, 193, 7, 0.45);
    transform: scale(1);
  }
  70% {
    color: #212121;
  }
  100% {
    box-shadow: 0 0 0 0 rgba(255, 193, 7, 0);
    transform: scale(1);
  }
}

@keyframes cell-flash-overlay {
  0% {
    opacity: 0.9;
  }
  30% {
    opacity: 0.55;
  }
  100% {
    opacity: 0;
  }
}
</style>
