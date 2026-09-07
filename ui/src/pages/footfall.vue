<template>
  <div class="footfall d-flex flex-column fill-height" style="overflow-y: auto">
    <!-- Header / controls -->
    <div class="d-flex align-center flex-wrap ga-3 pa-4 pb-2">
      <v-icon icon="mdi-walk" />
      <div class="text-h6 font-weight-medium">Footfall</div>
      <v-spacer />
      <v-select
        v-model="rangeMonths"
        density="compact"
        hide-details
        :items="RANGE_OPTIONS"
        style="max-width: 170px"
        variant="outlined"
      />
      <v-select
        v-model="selectedBranchIds"
        chips
        closable-chips
        density="compact"
        hide-details
        :items="branchOptions"
        label="Branches"
        multiple
        style="min-width: 220px; max-width: 340px"
        variant="outlined"
      >
        <template #prepend-item>
          <v-list-item title="All branches" @click="selectedBranchIds = []" />
          <v-divider />
        </template>
      </v-select>
    </div>

    <v-alert v-if="error" class="ff-notice mx-4 mb-2" density="compact" type="error" variant="tonal">
      {{ error }}
    </v-alert>

    <div v-if="loading" class="d-flex justify-center pa-12">
      <v-progress-circular color="primary" indeterminate />
    </div>

    <v-alert v-else-if="rows.length === 0" class="ff-notice mx-4" type="info" variant="tonal">
      No visitor counts in this range. Branches send these from the nav-drawer people counter in the POS.
    </v-alert>

    <v-container v-else class="pa-4 pt-2" fluid>
      <!-- KPIs -->
      <v-row dense>
        <v-col v-for="kpi in kpis" :key="kpi.label" cols="6" md="2">
          <v-card class="pa-3 d-flex flex-column justify-center fill-height" variant="tonal">
            <div class="d-flex align-center justify-space-between">
              <div class="text-h6 font-weight-medium">{{ kpi.value }}</div>
              <v-chip
                v-if="kpi.delta != null"
                :color="kpi.delta >= 0 ? 'primary' : 'error'"
                size="small"
                variant="tonal"
              >
                <v-icon :icon="kpi.delta >= 0 ? 'mdi-arrow-up' : 'mdi-arrow-down'" start />
                {{ Math.abs(kpi.delta) }}%
              </v-chip>
            </div>
            <div class="text-caption text-medium-emphasis mt-1">{{ kpi.label }}</div>
            <div v-if="kpi.hint" class="text-caption text-disabled">{{ kpi.hint }}</div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Trend + weekday -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Daily footfall — customers in</div>
            <div v-if="tooManyBranchesForTrend" class="chart-card__subtitle">
              {{ activeBranchIds.length }} branches selected — showing the combined total. Narrow the branch filter to compare individual lines.
            </div>
            <v-chart autoresize class="chart" :option="trendOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Average by weekday</div>
            <div class="chart-card__subtitle">Mean customers in per day of week over the range</div>
            <v-chart autoresize class="chart" :option="weekdayOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Month comparison + branch comparison -->
      <v-row>
        <v-col cols="12" md="7">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Month over month by branch</div>
            <v-chart autoresize class="chart" :option="monthlyOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="5">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Branch comparison — total customers in</div>
            <v-chart autoresize class="chart" :option="branchOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Month x branch table -->
      <v-row>
        <v-col cols="12">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Customers in — month × branch</div>
            <div class="table-scroll">
              <v-table density="compact">
                <thead>
                  <tr>
                    <th class="text-left">Month</th>
                    <th v-for="b in tableBranches" :key="b.id" class="text-right">{{ b.name }}</th>
                    <th class="text-right font-weight-bold">Total</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="m in months" :key="m">
                    <td>{{ monthLabel(m) }}</td>
                    <td v-for="b in tableBranches" :key="b.id" class="text-right">
                      {{ (matrix[m]?.[b.id] ?? 0).toLocaleString() }}
                    </td>
                    <td class="text-right font-weight-medium">{{ (monthTotals[m] ?? 0).toLocaleString() }}</td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr>
                    <td class="font-weight-bold">Total</td>
                    <td v-for="b in tableBranches" :key="b.id" class="text-right font-weight-bold">
                      {{ (branchTotals[b.id] ?? 0).toLocaleString() }}
                    </td>
                    <td class="text-right font-weight-bold">{{ grandTotalIn.toLocaleString() }}</td>
                  </tr>
                </tfoot>
              </v-table>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup>
  import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
  import VChart from 'vue-echarts'
  import { useTheme } from 'vuetify'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchVisitors } from '@/composables/useBranchVisitors'

  const theme = useTheme()
  const c = computed(() => theme.current.value.colors)
  const isDark = computed(() => theme.current.value.dark)
  const primary = computed(() => c.value.primary)
  const axisColor = computed(() => (isDark.value ? '#3a3a3a' : '#e0e0e0'))
  const textColor = computed(() => (isDark.value ? '#aaa' : '#666'))

  const catAxis = computed(() => ({
    axisLine: { lineStyle: { color: axisColor.value } },
    axisTick: { show: false },
    axisLabel: { color: textColor.value, fontSize: 11 },
  }))
  const valAxis = computed(() => ({
    axisLabel: { color: textColor.value, fontSize: 11 },
    splitLine: { lineStyle: { color: axisColor.value, type: 'dashed' } },
  }))
  const legendBase = computed(() => ({
    top: 0,
    right: 0,
    itemWidth: 10,
    itemHeight: 10,
    textStyle: { color: textColor.value, fontSize: 10 },
  }))

  const RANGE_OPTIONS = [
    { title: 'Last 3 months', value: 3 },
    { title: 'Last 6 months', value: 6 },
    { title: 'Last 12 months', value: 12 },
  ]
  // Above this many branches the per-branch trend lines stop being
  // readable (echarts would cycle its palette); fall back to one combined
  // line instead of inventing colours.
  const MAX_TREND_BRANCHES = 6

  const rangeMonths = ref(6)
  const selectedBranchIds = ref([]) // empty = all

  const { branches, fetchBranches } = useBranches()
  const { listVisitorStats } = useBranchVisitors()
  const { ensureConnected, onMessage } = useAdminSocket()

  const loading = ref(false)
  const error = ref('')
  // Raw rows for [prevFrom, to] -- current + one preceding period, split
  // client-side so the KPI deltas don't need a second request.
  const allDays = ref([])

  function ymd (d) {
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  }
  function firstOfMonth (year, monthIndex) {
    return new Date(year, monthIndex, 1)
  }

  // Current window: rangeMonths whole months ending with the current month,
  // through today. Previous window: the rangeMonths months immediately
  // before it.
  const windowBounds = computed(() => {
    const today = new Date()
    const start = firstOfMonth(today.getFullYear(), today.getMonth() - (rangeMonths.value - 1))
    const prevStart = firstOfMonth(start.getFullYear(), start.getMonth() - rangeMonths.value)
    return { from: ymd(start), to: ymd(today), prevFrom: ymd(prevStart) }
  })

  function fetchWindow () {
    const { prevFrom, to } = windowBounds.value
    return listVisitorStats({ from: prevFrom, to })
  }

  async function load () {
    loading.value = true
    error.value = ''
    try {
      await fetchBranches()
      allDays.value = await fetchWindow()
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  }

  // Background refetch -- no full-page spinner, and it keeps the last good
  // data on failure. Reconciles the optimistic updates below against the
  // server (other branches, out-of-order or missed pushes).
  const refreshing = ref(false)
  async function backgroundRefresh () {
    if (refreshing.value) return
    refreshing.value = true
    try {
      allDays.value = await fetchWindow()
      error.value = ''
    } catch { /* keep showing the last successful fetch */ } finally {
      refreshing.value = false
    }
  }

  // Apply a live counter press to the in-memory rollup immediately, so the
  // charts and KPIs move the instant a customer is counted instead of
  // waiting for the debounced refetch. message.amount is the authoritative
  // running gross tally for that branch/day/direction, so set (clamped
  // upward) rather than blindly increment -- a duplicate push is then a
  // no-op.
  function applyLiveEvent (message) {
    const day = message.label
    const branchId = Number(message.branch_id)
    if (!day || !branchId) return
    const key = message.kind === 'out' ? 'out_count' : 'in_count'
    const n = Number(message.amount) || 0
    const row = allDays.value.find(r => r.branch_id === branchId && r.day === day)
    if (row) {
      row[key] = Math.max(row[key] ?? 0, n)
    } else {
      allDays.value = [
        ...allDays.value,
        { branch_id: branchId, day, in_count: key === 'in_count' ? n : 0, out_count: key === 'out_count' ? n : 0 },
      ]
    }
  }

  // Trailing debounce that can't be starved: once a refetch is pending it
  // stays pending (~4s out) no matter how many more presses arrive.
  let refreshTimer = null
  function scheduleRefresh () {
    if (refreshTimer) return
    refreshTimer = setTimeout(() => {
      refreshTimer = null
      backgroundRefresh()
    }, 4000)
  }
  let unsubscribe = null

  onMounted(() => {
    load()
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (message.type !== 'branch_visitor_event') return
      applyLiveEvent(message)
      scheduleRefresh()
    })
  })
  onUnmounted(() => {
    clearTimeout(refreshTimer)
    unsubscribe?.()
  })

  // Reload when the range changes (the fetched window depends on it); the
  // branch filter is applied client-side so it needs no refetch.
  watch(rangeMonths, load)

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }

  const branchOptions = computed(() =>
    [...branches.value]
      .toSorted((a, b) => a.name.localeCompare(b.name))
      .map(b => ({ title: b.name, value: b.id })),
  )

  // Branch ids that actually have data in the current window, intersected
  // with the user's selection (empty selection = all of them).
  const branchIdsWithData = computed(() => {
    const { from } = windowBounds.value
    const ids = new Set()
    for (const r of allDays.value) {
      if (r.day >= from) ids.add(r.branch_id)
    }
    return [...ids]
  })
  const activeBranchIds = computed(() => {
    const sel = selectedBranchIds.value
    if (sel.length === 0) return branchIdsWithData.value
    return branchIdsWithData.value.filter(id => sel.includes(id))
  })

  // Current-window rows for the active branches.
  const rows = computed(() => {
    const { from } = windowBounds.value
    const active = new Set(activeBranchIds.value)
    return allDays.value.filter(r => r.day >= from && active.has(r.branch_id))
  })
  const prevRows = computed(() => {
    const { from } = windowBounds.value
    const active = new Set(activeBranchIds.value)
    return allDays.value.filter(r => r.day < from && active.has(r.branch_id))
  })

  const sum = (arr, key) => arr.reduce((t, r) => t + (r[key] ?? 0), 0)

  function monthKey (day) {
    return day.slice(0, 7)
  }
  function monthLabel (key) {
    const [y, m] = key.split('-')
    return new Date(Number(y), Number(m) - 1, 1)
      .toLocaleDateString(undefined, { month: 'short', year: '2-digit' })
  }

  // Every month in the current window, oldest first, even if it has no rows
  // -- so month-over-month comparisons don't silently skip quiet months.
  const months = computed(() => {
    const { from, to } = windowBounds.value
    const out = []
    const d = new Date(`${from}T00:00:00`)
    const end = new Date(`${to}T00:00:00`)
    while (d <= end) {
      out.push(`${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`)
      d.setMonth(d.getMonth() + 1)
    }
    return out
  })

  const days = computed(() => [...new Set(rows.value.map(r => r.day))].toSorted())

  // --- KPIs ---------------------------------------------------------------
  const totalIn = computed(() => sum(rows.value, 'in_count'))
  const totalOut = computed(() => sum(rows.value, 'out_count'))
  const prevTotalIn = computed(() => sum(prevRows.value, 'in_count'))

  const busiestDay = computed(() => {
    const byDay = {}
    for (const r of rows.value) byDay[r.day] = (byDay[r.day] ?? 0) + r.in_count
    let best = null
    for (const [day, n] of Object.entries(byDay)) {
      if (!best || n > best.n) best = { day, n }
    }
    return best
  })
  const busiestBranch = computed(() => {
    const byBranch = {}
    for (const r of rows.value) byBranch[r.branch_id] = (byBranch[r.branch_id] ?? 0) + r.in_count
    let best = null
    for (const [id, n] of Object.entries(byBranch)) {
      if (!best || n > best.n) best = { id: Number(id), n }
    }
    return best
  })

  const kpis = computed(() => {
    const activeDayCount = days.value.length || 1
    const delta = prevTotalIn.value > 0
      ? Math.round(((totalIn.value - prevTotalIn.value) / prevTotalIn.value) * 100)
      : null
    return [
      { label: 'Customers in', value: totalIn.value.toLocaleString(), delta, hint: 'vs previous period' },
      { label: 'Customers out', value: totalOut.value.toLocaleString() },
      { label: 'Net', value: (totalIn.value - totalOut.value).toLocaleString() },
      { label: 'Avg in / day', value: Math.round(totalIn.value / activeDayCount).toLocaleString() },
      {
        label: 'Busiest day',
        value: busiestDay.value ? String(busiestDay.value.n) : '—',
        hint: busiestDay.value ? new Date(`${busiestDay.value.day}T00:00:00`).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) : '',
      },
      {
        label: 'Busiest branch',
        value: busiestBranch.value ? String(busiestBranch.value.n) : '—',
        hint: busiestBranch.value ? branchName(busiestBranch.value.id) : '',
      },
    ]
  })

  // --- charts -----------------------------------------------------------
  const tooManyBranchesForTrend = computed(() => activeBranchIds.value.length > MAX_TREND_BRANCHES)

  const trendOption = computed(() => {
    const xs = days.value
    const byDayBranch = {}
    for (const r of rows.value) {
      byDayBranch[r.branch_id] ??= {}
      byDayBranch[r.branch_id][r.day] = r.in_count
    }

    let series
    if (tooManyBranchesForTrend.value) {
      const byDay = {}
      for (const r of rows.value) byDay[r.day] = (byDay[r.day] ?? 0) + r.in_count
      series = [{
        name: 'All selected branches',
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2 },
        itemStyle: { color: primary.value },
        areaStyle: { opacity: 0.08 },
        data: xs.map(d => byDay[d] ?? 0),
      }]
    } else {
      series = activeBranchIds.value.map(id => ({
        name: branchName(id),
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2 },
        data: xs.map(d => byDayBranch[id]?.[d] ?? 0),
      }))
    }

    return {
      grid: { top: series.length > 1 ? 34 : 20, right: 12, bottom: 28, left: 44 },
      tooltip: { trigger: 'axis', appendTo: 'body' },
      legend: series.length > 1 ? { ...legendBase.value, data: series.map(s => s.name) } : undefined,
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: xs.map(d => d.slice(5)),
        ...catAxis.value,
      },
      yAxis: { type: 'value', minInterval: 1, ...valAxis.value },
      series,
    }
  })

  const weekdayOption = computed(() => {
    const LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
    const totals = Array.from({ length: 7 }, () => 0)
    const counts = Array.from({ length: 7 }, () => 0)
    const byDay = {}
    for (const r of rows.value) byDay[r.day] = (byDay[r.day] ?? 0) + r.in_count
    for (const [day, n] of Object.entries(byDay)) {
      // getDay(): 0 = Sunday -> map to a Monday-first index.
      const idx = (new Date(`${day}T00:00:00`).getDay() + 6) % 7
      totals[idx] += n
      counts[idx] += 1
    }
    const avg = totals.map((t, i) => (counts[i] ? Math.round(t / counts[i]) : 0))
    return {
      grid: { top: 16, right: 12, bottom: 24, left: 36 },
      tooltip: { trigger: 'axis', appendTo: 'body' },
      xAxis: { type: 'category', data: LABELS, ...catAxis.value },
      yAxis: { type: 'value', minInterval: 1, ...valAxis.value },
      series: [{
        type: 'bar',
        barWidth: '55%',
        itemStyle: { color: primary.value, borderRadius: [3, 3, 0, 0] },
        data: avg,
      }],
    }
  })

  const monthlyOption = computed(() => {
    const ids = activeBranchIds.value
    const perMonthBranch = {}
    for (const r of rows.value) {
      const mk = monthKey(r.day)
      perMonthBranch[mk] ??= {}
      perMonthBranch[mk][r.branch_id] = (perMonthBranch[mk][r.branch_id] ?? 0) + r.in_count
    }
    const series = ids.map(id => ({
      name: branchName(id),
      type: 'bar',
      data: months.value.map(m => perMonthBranch[m]?.[id] ?? 0),
    }))
    return {
      grid: { top: series.length > 1 ? 34 : 18, right: 12, bottom: 26, left: 44 },
      tooltip: { trigger: 'axis', appendTo: 'body', axisPointer: { type: 'shadow' } },
      legend: series.length > 1 ? { ...legendBase.value, data: series.map(s => s.name) } : undefined,
      xAxis: { type: 'category', data: months.value.map(monthLabel), ...catAxis.value },
      yAxis: { type: 'value', minInterval: 1, ...valAxis.value },
      series,
    }
  })

  const branchOption = computed(() => {
    const totals = activeBranchIds.value
      .map(id => ({ name: branchName(id), value: rows.value.filter(r => r.branch_id === id).reduce((t, r) => t + r.in_count, 0) }))
      .toSorted((a, b) => a.value - b.value)
    return {
      grid: { top: 8, right: 40, bottom: 24, left: 8, containLabel: true },
      tooltip: { trigger: 'axis', appendTo: 'body', axisPointer: { type: 'shadow' } },
      xAxis: { type: 'value', minInterval: 1, ...valAxis.value },
      yAxis: { type: 'category', data: totals.map(t => t.name), ...catAxis.value },
      series: [{
        type: 'bar',
        barWidth: '55%',
        label: { show: true, position: 'right', color: textColor.value, fontSize: 11 },
        itemStyle: { color: primary.value, borderRadius: [0, 3, 3, 0] },
        data: totals.map(t => t.value),
      }],
    }
  })

  // --- month x branch table -------------------------------------------
  const tableBranches = computed(() =>
    activeBranchIds.value
      .map(id => ({ id, name: branchName(id) }))
      .toSorted((a, b) => a.name.localeCompare(b.name)),
  )
  const matrix = computed(() => {
    const out = {}
    for (const r of rows.value) {
      const mk = monthKey(r.day)
      out[mk] ??= {}
      out[mk][r.branch_id] = (out[mk][r.branch_id] ?? 0) + r.in_count
    }
    return out
  })
  const monthTotals = computed(() => {
    const out = {}
    for (const m of months.value) {
      out[m] = tableBranches.value.reduce((t, b) => t + (matrix.value[m]?.[b.id] ?? 0), 0)
    }
    return out
  })
  const branchTotals = computed(() => {
    const out = {}
    for (const b of tableBranches.value) {
      out[b.id] = rows.value.filter(r => r.branch_id === b.id).reduce((t, r) => t + r.in_count, 0)
    }
    return out
  })
  const grandTotalIn = computed(() => totalIn.value)
</script>

<style scoped>
/* v-alert defaults to flex: 1 1 inside a flex column, which stretched the
   empty / error notice to the whole page -- pin it to its content height. */
.ff-notice {
  flex: 0 0 auto;
}
/* v-row stretches its columns, so height:100% makes every card's bordered
   box the same height as the tallest in its row. The chart keeps a FIXED
   height -- a flex-grown height feeds back into echarts' autoresize and the
   card grows without bound. */
.chart-card {
  padding: 12px 14px 4px;
  margin-bottom: 8px;
  border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  border-radius: 10px;
  height: 100%;
}
.chart-card__title {
  font-size: 0.9rem;
  font-weight: 600;
  margin-bottom: 2px;
}
.chart-card__subtitle {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
  margin-bottom: 4px;
}
.chart {
  height: 280px;
  width: 100%;
}
.table-scroll {
  overflow-x: auto;
  padding-bottom: 8px;
}
</style>
