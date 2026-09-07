<template>
  <div class="conversion d-flex flex-column fill-height" style="overflow-y: auto">
    <!-- Header / controls -- same shape as the Footfall page it extends -->
    <div class="d-flex align-center flex-wrap ga-3 pa-4 pb-2">
      <v-icon icon="mdi-percent-outline" />
      <div class="text-h6 font-weight-medium">Conversion</div>
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

    <v-alert
      v-if="error"
      class="cv-notice mx-4 mb-2"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ error }}
    </v-alert>

    <div v-if="loading" class="d-flex justify-center pa-12">
      <v-progress-circular color="primary" indeterminate />
    </div>

    <v-alert v-else-if="rows.length === 0" class="cv-notice mx-4" type="info" variant="tonal">
      No footfall or sales in this range.
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

      <!-- Conversion trend + weekday profile -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Conversion rate by month</div>
            <div v-if="tooManyBranchesForTrend" class="chart-card__subtitle">
              {{ activeBranchIds.length }} branches selected — showing the combined rate. Narrow the branch filter to compare individual lines.
            </div>
            <v-chart autoresize class="chart" :option="monthlyOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Conversion by weekday</div>
            <v-chart autoresize class="chart" :option="weekdayOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Funnel (visitors vs invoices) + branch comparison -->
      <v-row>
        <v-col cols="12" md="7">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Visitors vs invoices by month</div>
            <v-chart autoresize class="chart" :option="funnelOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="5">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Branch comparison — conversion rate</div>
            <v-chart autoresize class="chart" :option="branchOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Month x branch conversion table -->
      <v-row>
        <v-col cols="12">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Conversion rate — month × branch</div>
            <div class="table-scroll">
              <v-table density="compact">
                <thead>
                  <tr>
                    <th class="text-left">Month</th>
                    <th v-for="b in tableBranches" :key="b.id" class="text-right">{{ b.name }}</th>
                    <th class="text-right font-weight-bold">Overall</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="m in months" :key="m">
                    <td>{{ monthLabel(m) }}</td>
                    <td v-for="b in tableBranches" :key="b.id" class="text-right">
                      {{ pctOrDash(matrix[m]?.[b.id]) }}
                    </td>
                    <td class="text-right font-weight-medium">{{ pctOrDash(monthOverall[m]) }}</td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr>
                    <td class="font-weight-bold">Overall</td>
                    <td v-for="b in tableBranches" :key="b.id" class="text-right font-weight-bold">
                      {{ pctOrDash(branchOverall[b.id]) }}
                    </td>
                    <td class="text-right font-weight-bold">{{ pctOrDash(overallRate) }}</td>
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
  import { useBranchInvoices } from '@/composables/useBranchInvoices'
  import { useBranchVisitors } from '@/composables/useBranchVisitors'

  // --- echarts chrome (identical helpers to footfall.vue) --------------
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
  // Past this many branches the per-branch trend lines just cycle echarts'
  // palette and stop being readable -- collapse to one combined rate line
  // instead, same threshold Footfall uses.
  const MAX_TREND_BRANCHES = 6

  const rangeMonths = ref(6)
  const selectedBranchIds = ref([]) // empty = all

  const { branches, fetchBranches } = useBranches()
  const { listVisitorStats } = useBranchVisitors()
  const { listSalesStats } = useBranchInvoices()
  const { ensureConnected, onMessage } = useAdminSocket()

  const loading = ref(false)
  const error = ref('')
  // Both sides of the funnel over [prevFrom, to] -- current window plus one
  // preceding window, so the conversion-rate KPI can show a period delta
  // without a second round of requests. Kept as the raw API rows; merged
  // per (branch, day) in `merged` below.
  const visitorDays = ref([])
  const salesDays = ref([])

  function ymd (d) {
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  }
  function firstOfMonth (year, monthIndex) {
    return new Date(year, monthIndex, 1)
  }

  const windowBounds = computed(() => {
    const today = new Date()
    const start = firstOfMonth(today.getFullYear(), today.getMonth() - (rangeMonths.value - 1))
    const prevStart = firstOfMonth(start.getFullYear(), start.getMonth() - rangeMonths.value)
    return { from: ymd(start), to: ymd(today), prevFrom: ymd(prevStart) }
  })

  function fetchWindow () {
    const { prevFrom, to } = windowBounds.value
    return Promise.all([
      listVisitorStats({ from: prevFrom, to }),
      listSalesStats({ from: prevFrom, to }),
    ])
  }

  async function load () {
    loading.value = true
    error.value = ''
    try {
      await fetchBranches()
      const [visitors, sales] = await fetchWindow()
      visitorDays.value = visitors
      salesDays.value = sales
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  }

  // Background refetch on a live event -- no spinner, last good data kept on
  // failure. No optimistic in-memory merge here (Footfall has one for its
  // counter); a debounced refetch is enough for a rate that moves slowly.
  const refreshing = ref(false)
  async function backgroundRefresh () {
    if (refreshing.value) return
    refreshing.value = true
    try {
      const [visitors, sales] = await fetchWindow()
      visitorDays.value = visitors
      salesDays.value = sales
      error.value = ''
    } catch { /* keep the last successful fetch */ } finally {
      refreshing.value = false
    }
  }

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
      // A counter press moves the denominator, a synced invoice the
      // numerator -- either should nudge the page.
      if (message.type === 'branch_visitor_event' || message.type === 'branch_invoice_created') {
        scheduleRefresh()
      }
    })
  })
  onUnmounted(() => {
    clearTimeout(refreshTimer)
    unsubscribe?.()
  })

  watch(rangeMonths, load)

  function branchName (id) {
    return branches.value.find(b => b.id === id)?.name ?? `Branch #${id}`
  }
  const branchOptions = computed(() =>
    [...branches.value]
      .toSorted((a, b) => a.name.localeCompare(b.name))
      .map(b => ({ title: b.name, value: b.id })),
  )

  function monthKey (day) {
    return day.slice(0, 7)
  }
  function monthLabel (key) {
    const [y, m] = key.split('-')
    return new Date(Number(y), Number(m) - 1, 1)
      .toLocaleDateString(undefined, { month: 'short', year: '2-digit' })
  }
  function money (cents) {
    return `$${Math.round((cents ?? 0) / 100).toLocaleString()}`
  }
  // ratio -> "12.3%"; null/blank denominator -> em dash.
  function pct (x) {
    return `${(x * 100).toFixed(1)}%`
  }
  function pctOrDash (x) {
    return x == null ? '—' : pct(x)
  }
  const rate = (invoices, visitors) => (visitors > 0 ? invoices / visitors : null)

  // Every month in the current window, oldest first, even the quiet ones.
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

  // --- merge the two feeds into one row per (branch, day) -------------
  // key: "branchId|day" -> { branch_id, day, visitors, invoices, revenue }
  const merged = computed(() => {
    const map = new Map()
    const touch = (branchId, day) => {
      const k = `${branchId}|${day}`
      let row = map.get(k)
      if (!row) {
        row = { branch_id: branchId, day, visitors: 0, invoices: 0, revenue: 0 }
        map.set(k, row)
      }
      return row
    }
    for (const v of visitorDays.value) touch(v.branch_id, v.day).visitors += v.in_count ?? 0
    for (const s of salesDays.value) {
      const row = touch(s.branch_id, s.day)
      row.invoices += s.invoice_count ?? 0
      row.revenue += s.revenue ?? 0
    }
    return [...map.values()]
  })

  const branchIdsWithData = computed(() => {
    const { from } = windowBounds.value
    const ids = new Set()
    for (const r of merged.value) {
      if (r.day >= from) ids.add(r.branch_id)
    }
    return [...ids]
  })
  const activeBranchIds = computed(() => {
    const sel = selectedBranchIds.value
    if (sel.length === 0) return branchIdsWithData.value
    return branchIdsWithData.value.filter(id => sel.includes(id))
  })

  const rows = computed(() => {
    const { from } = windowBounds.value
    const active = new Set(activeBranchIds.value)
    return merged.value.filter(r => r.day >= from && active.has(r.branch_id))
  })
  const prevRows = computed(() => {
    const { from } = windowBounds.value
    const active = new Set(activeBranchIds.value)
    return merged.value.filter(r => r.day < from && active.has(r.branch_id))
  })

  const sum = (arr, key) => arr.reduce((t, r) => t + (r[key] ?? 0), 0)

  // --- KPIs ---------------------------------------------------------------
  const totalVisitors = computed(() => sum(rows.value, 'visitors'))
  const totalInvoices = computed(() => sum(rows.value, 'invoices'))
  const totalRevenue = computed(() => sum(rows.value, 'revenue'))
  const overallRate = computed(() => rate(totalInvoices.value, totalVisitors.value))
  const prevRate = computed(() => rate(sum(prevRows.value, 'invoices'), sum(prevRows.value, 'visitors')))

  const kpis = computed(() => {
    // Delta is in percentage points of the rate, not a percent-of-percent --
    // "conversion went from 18% to 21%" reads as +3, not +17.
    const delta = (overallRate.value != null && prevRate.value != null)
      ? Math.round((overallRate.value - prevRate.value) * 100)
      : null
    return [
      {
        label: 'Conversion rate',
        value: overallRate.value == null ? '—' : pct(overallRate.value),
        delta,
        hint: 'invoices ÷ visitors',
      },
      { label: 'Visitors in', value: totalVisitors.value.toLocaleString() },
      { label: 'Invoices', value: totalInvoices.value.toLocaleString() },
      {
        label: 'Avg invoice',
        value: totalInvoices.value > 0 ? money(totalRevenue.value / totalInvoices.value) : '—',
        hint: 'revenue ÷ invoices',
      },
      {
        label: 'Revenue / visitor',
        value: totalVisitors.value > 0 ? money(totalRevenue.value / totalVisitors.value) : '—',
      },
      { label: 'Revenue', value: money(totalRevenue.value) },
    ]
  })

  // --- charts -----------------------------------------------------------
  const tooManyBranchesForTrend = computed(() => activeBranchIds.value.length > MAX_TREND_BRANCHES)

  // Per-branch (or combined) conversion rate per month. Rate of sums, not a
  // mean of daily ratios -- a day with 1 visitor shouldn't swing the month.
  const monthlyOption = computed(() => {
    const pctFmt = { formatter: v => `${Math.round(v)}%` }

    let series
    if (tooManyBranchesForTrend.value) {
      const byMonth = {}
      for (const r of rows.value) {
        const mk = monthKey(r.day)
        byMonth[mk] ??= { inv: 0, vis: 0 }
        byMonth[mk].inv += r.invoices
        byMonth[mk].vis += r.visitors
      }
      series = [{
        name: 'All selected branches',
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2 },
        itemStyle: { color: primary.value },
        areaStyle: { opacity: 0.08 },
        data: months.value.map(m => {
          const r = rate(byMonth[m]?.inv ?? 0, byMonth[m]?.vis ?? 0)
          return r == null ? null : +(r * 100).toFixed(1)
        }),
      }]
    } else {
      const byMonthBranch = {}
      for (const r of rows.value) {
        const mk = monthKey(r.day)
        byMonthBranch[r.branch_id] ??= {}
        byMonthBranch[r.branch_id][mk] ??= { inv: 0, vis: 0 }
        byMonthBranch[r.branch_id][mk].inv += r.invoices
        byMonthBranch[r.branch_id][mk].vis += r.visitors
      }
      series = activeBranchIds.value.map(id => ({
        name: branchName(id),
        type: 'line',
        smooth: true,
        showSymbol: false,
        connectNulls: true,
        lineStyle: { width: 2 },
        data: months.value.map(m => {
          const cell = byMonthBranch[id]?.[m]
          const r = rate(cell?.inv ?? 0, cell?.vis ?? 0)
          return r == null ? null : +(r * 100).toFixed(1)
        }),
      }))
    }

    return {
      grid: { top: series.length > 1 ? 34 : 20, right: 12, bottom: 26, left: 44 },
      tooltip: {
        trigger: 'axis',
        appendTo: 'body',
        valueFormatter: v => (v == null ? '—' : `${v}%`),
      },
      legend: series.length > 1 ? { ...legendBase.value, data: series.map(s => s.name) } : undefined,
      xAxis: { type: 'category', data: months.value.map(m => monthLabel(m)), ...catAxis.value },
      yAxis: { type: 'value', axisLabel: { ...valAxis.value.axisLabel, ...pctFmt }, splitLine: valAxis.value.splitLine },
      series,
    }
  })

  const weekdayOption = computed(() => {
    const LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
    const inv = Array.from({ length: 7 }, () => 0)
    const vis = Array.from({ length: 7 }, () => 0)
    for (const r of rows.value) {
      const idx = (new Date(`${r.day}T00:00:00`).getDay() + 6) % 7
      inv[idx] += r.invoices
      vis[idx] += r.visitors
    }
    const data = inv.map((n, i) => {
      const rt = rate(n, vis[i])
      return rt == null ? null : +(rt * 100).toFixed(1)
    })
    return {
      grid: { top: 16, right: 12, bottom: 24, left: 40 },
      tooltip: { trigger: 'axis', appendTo: 'body', valueFormatter: v => (v == null ? '—' : `${v}%`) },
      xAxis: { type: 'category', data: LABELS, ...catAxis.value },
      yAxis: { type: 'value', axisLabel: { ...valAxis.value.axisLabel, formatter: v => `${Math.round(v)}%` }, splitLine: valAxis.value.splitLine },
      series: [{
        type: 'bar',
        barWidth: '55%',
        itemStyle: { color: primary.value, borderRadius: [3, 3, 0, 0] },
        data,
      }],
    }
  })

  // Bars = visitors, line = invoices, on their own axes -- the raw funnel
  // behind the rate, combined across the selected branches.
  const funnelOption = computed(() => {
    const byMonth = {}
    for (const r of rows.value) {
      const mk = monthKey(r.day)
      byMonth[mk] ??= { vis: 0, inv: 0 }
      byMonth[mk].vis += r.visitors
      byMonth[mk].inv += r.invoices
    }
    return {
      // Axis names are dropped -- the legend already labels bar vs line, and
      // a right-side yAxis `name` renders exactly where the legend sits.
      grid: { top: 30, right: 46, bottom: 26, left: 48 },
      tooltip: { trigger: 'axis', appendTo: 'body', axisPointer: { type: 'shadow' } },
      legend: { ...legendBase.value, data: ['Visitors', 'Invoices'] },
      xAxis: { type: 'category', data: months.value.map(m => monthLabel(m)), ...catAxis.value },
      yAxis: [
        { type: 'value', minInterval: 1, ...valAxis.value },
        { type: 'value', minInterval: 1, splitLine: { show: false }, axisLabel: valAxis.value.axisLabel },
      ],
      series: [
        {
          name: 'Visitors',
          type: 'bar',
          barWidth: '45%',
          itemStyle: { color: primary.value, borderRadius: [3, 3, 0, 0] },
          data: months.value.map(m => byMonth[m]?.vis ?? 0),
        },
        {
          name: 'Invoices',
          type: 'line',
          smooth: true,
          showSymbol: false,
          yAxisIndex: 1,
          lineStyle: { width: 2 },
          data: months.value.map(m => byMonth[m]?.inv ?? 0),
        },
      ],
    }
  })

  const branchOption = computed(() => {
    const totals = activeBranchIds.value
      .map(id => {
        const br = rows.value.filter(r => r.branch_id === id)
        return { name: branchName(id), value: rate(sum(br, 'invoices'), sum(br, 'visitors')) }
      })
      .filter(t => t.value != null)
      .map(t => ({ name: t.name, value: +(t.value * 100).toFixed(1) }))
      .toSorted((a, b) => a.value - b.value)
    return {
      grid: { top: 8, right: 48, bottom: 24, left: 8, containLabel: true },
      tooltip: { trigger: 'axis', appendTo: 'body', axisPointer: { type: 'shadow' }, valueFormatter: v => `${v}%` },
      xAxis: { type: 'value', axisLabel: { ...valAxis.value.axisLabel, formatter: v => `${Math.round(v)}%` }, splitLine: valAxis.value.splitLine },
      yAxis: { type: 'category', data: totals.map(t => t.name), ...catAxis.value },
      series: [{
        type: 'bar',
        barWidth: '55%',
        label: { show: true, position: 'right', color: textColor.value, fontSize: 11, formatter: p => `${p.value}%` },
        itemStyle: { color: primary.value, borderRadius: [0, 3, 3, 0] },
        data: totals.map(t => t.value),
      }],
    }
  })

  // --- month x branch table -----------------------------------------
  const tableBranches = computed(() =>
    activeBranchIds.value
      .map(id => ({ id, name: branchName(id) }))
      .toSorted((a, b) => a.name.localeCompare(b.name)),
  )
  // matrix[month][branchId] = rate (or absent). Also feeds the row / column
  // overalls, each its own rate-of-sums rather than an average of cells.
  const aggregates = computed(() => {
    const cell = {} // `${m}|${id}` -> { inv, vis }
    const monthAgg = {}
    const branchAgg = {}
    for (const r of rows.value) {
      const m = monthKey(r.day)
      const ck = `${m}|${r.branch_id}`
      cell[ck] ??= { inv: 0, vis: 0 }
      cell[ck].inv += r.invoices
      cell[ck].vis += r.visitors
      monthAgg[m] ??= { inv: 0, vis: 0 }
      monthAgg[m].inv += r.invoices
      monthAgg[m].vis += r.visitors
      branchAgg[r.branch_id] ??= { inv: 0, vis: 0 }
      branchAgg[r.branch_id].inv += r.invoices
      branchAgg[r.branch_id].vis += r.visitors
    }
    return { cell, monthAgg, branchAgg }
  })
  const matrix = computed(() => {
    const out = {}
    for (const [k, v] of Object.entries(aggregates.value.cell)) {
      const [m, id] = k.split('|')
      out[m] ??= {}
      out[m][Number(id)] = rate(v.inv, v.vis)
    }
    return out
  })
  const monthOverall = computed(() => {
    const out = {}
    for (const [m, v] of Object.entries(aggregates.value.monthAgg)) out[m] = rate(v.inv, v.vis)
    return out
  })
  const branchOverall = computed(() => {
    const out = {}
    for (const [id, v] of Object.entries(aggregates.value.branchAgg)) out[Number(id)] = rate(v.inv, v.vis)
    return out
  })
</script>

<style scoped>
/* v-alert defaults to flex: 1 1 inside a flex column, which stretched the
   empty / error notice to the whole page -- pin it to its content height. */
.cv-notice {
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
