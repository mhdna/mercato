<template>
  <div class="loyalty d-flex flex-column fill-height" style="overflow-y: auto">
    <!-- Header / controls -- same shape as the Footfall / Conversion pages. -->
    <div class="d-flex align-center flex-wrap ga-3 pa-4 pb-2">
      <v-icon icon="mdi-trophy" />
      <div class="text-h6 font-weight-medium">Loyalty</div>
      <v-spacer />
      <v-btn-toggle
        v-model="clientType"
        density="compact"
        divided
        mandatory
        variant="outlined"
      >
        <v-btn v-for="opt in clientTypeOptions" :key="opt.value" :value="opt.value">
          {{ opt.title }}
        </v-btn>
      </v-btn-toggle>
    </div>

    <!-- Tier ladder: the thresholds every number on this page is bucketed by.
      Edited from the Clients page; shared across retail and wholesale. -->
    <div class="d-flex flex-wrap align-center ga-2 px-4 pb-2">
      <span class="text-caption text-medium-emphasis me-1">Tiers</span>
      <v-chip
        v-for="t in tierLadder"
        :key="t.id"
        class="text-white"
        :color="t.color"
        size="small"
        variant="flat"
      >
        {{ t.label }} · ${{ t.minAmount.toLocaleString() }}+
      </v-chip>
    </div>

    <v-alert
      v-if="error"
      class="ly-notice mx-4 mb-2"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ error }}
      <template #append>
        <v-btn size="small" text="Retry" variant="text" @click="load" />
      </template>
    </v-alert>

    <div v-if="loading" class="d-flex flex-column align-center ga-2 pa-12">
      <v-progress-circular color="primary" indeterminate />
      <span class="text-caption text-medium-emphasis">
        Loading clients… {{ loadedCount.toLocaleString() }}<template v-if="totalClients"> / {{ totalClients.toLocaleString() }}</template>
      </span>
    </div>

    <v-alert v-else-if="rows.length === 0" class="ly-notice mx-4" type="info" variant="tonal">
      No clients to analyse yet.
    </v-alert>

    <v-container v-else class="pa-4 pt-2" fluid>
      <v-alert
        v-if="truncated"
        class="mb-3"
        density="compact"
        type="info"
        variant="tonal"
      >
        Showing the top {{ rows.length.toLocaleString() }} of {{ totalClients.toLocaleString() }} clients by spend.
        The lowest tier's count is understated by the untracked tail.
      </v-alert>

      <!-- KPIs -->
      <v-row dense>
        <v-col v-for="kpi in kpis" :key="kpi.label" cols="6" md="2">
          <v-card class="pa-3 d-flex flex-column justify-center fill-height" variant="tonal">
            <div class="text-h6 font-weight-medium">{{ kpi.value }}</div>
            <div class="text-caption text-medium-emphasis mt-1">{{ kpi.label }}</div>
            <div v-if="kpi.hint" class="text-caption text-disabled">{{ kpi.hint }}</div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Members per tier + revenue share -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Members by tier</div>
            <div class="chart-card__subtitle">How many clients have reached each tier</div>
            <v-chart autoresize class="chart" :option="membersByTierOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Revenue share by tier</div>
            <div class="chart-card__subtitle">Lifetime spend, split by current tier</div>
            <v-chart autoresize class="chart" :option="revenueShareOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- New members trend + points liability -->
      <v-row>
        <v-col cols="12" md="8">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">New members per month</div>
            <div class="chart-card__subtitle">Sign-ups over the trailing 12 months</div>
            <v-chart autoresize class="chart" :option="newMembersOption" />
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Points outstanding by tier</div>
            <div class="chart-card__subtitle">Unredeemed balance held in each tier</div>
            <v-chart autoresize class="chart" :option="pointsByTierOption" />
          </v-card>
        </v-col>
      </v-row>

      <!-- Tier breakdown table + composition bars -->
      <v-row>
        <v-col cols="12" md="7">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Tier breakdown</div>
            <div class="table-scroll">
              <v-table density="compact">
                <thead>
                  <tr>
                    <th class="text-left">Tier</th>
                    <th class="text-right">Min spend</th>
                    <th class="text-right">Members</th>
                    <th class="text-right">% of base</th>
                    <th class="text-right">Revenue</th>
                    <th class="text-right">% of rev.</th>
                    <th class="text-right">Avg spend</th>
                    <th class="text-right">Points</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="t in tierStatsDesc" :key="t.id">
                    <td>
                      <span class="tier-dot" :style="{ backgroundColor: t.color }" />
                      {{ t.label }}
                    </td>
                    <td class="text-right">${{ t.minAmount.toLocaleString() }}</td>
                    <td class="text-right font-weight-medium">{{ t.members.toLocaleString() }}</td>
                    <td class="text-right">{{ pct(t.baseShare) }}</td>
                    <td class="text-right">${{ money(t.revenue) }}</td>
                    <td class="text-right">{{ pct(t.revShare) }}</td>
                    <td class="text-right">${{ money(t.avgSpend) }}</td>
                    <td class="text-right">{{ t.points.toLocaleString() }}</td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr>
                    <td class="font-weight-bold">All</td>
                    <td />
                    <td class="text-right font-weight-bold">{{ totalMembers.toLocaleString() }}</td>
                    <td class="text-right font-weight-bold">100%</td>
                    <td class="text-right font-weight-bold">${{ money(totalRevenue) }}</td>
                    <td class="text-right font-weight-bold">100%</td>
                    <td class="text-right font-weight-bold">
                      ${{ money(totalMembers ? totalRevenue / totalMembers : 0) }}
                    </td>
                    <td class="text-right font-weight-bold">{{ totalPoints.toLocaleString() }}</td>
                  </tr>
                </tfoot>
              </v-table>
            </div>
          </v-card>
        </v-col>

        <v-col cols="12" md="5">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Tier composition</div>
            <div class="chart-card__subtitle">Share of the member base in each tier</div>

            <div class="mt-3">
              <div v-for="t in tierStats" :key="t.id" class="comp-row">
                <div class="comp-label">
                  <span class="tier-dot" :style="{ backgroundColor: t.color }" />
                  {{ t.label }}
                </div>
                <v-progress-linear
                  class="comp-bar"
                  :color="t.color"
                  height="14"
                  :model-value="t.baseShare * 100"
                  rounded
                />
                <div class="comp-value">{{ t.members.toLocaleString() }} · {{ pct(t.baseShare) }}</div>
              </div>
            </div>

            <v-divider class="my-4" />

            <div class="text-caption text-medium-emphasis mb-1">
              Members past {{ entryTier?.label }}
            </div>
            <v-progress-linear
              color="primary"
              height="18"
              :model-value="advancedShare * 100"
              rounded
            >
              <span class="text-caption font-weight-medium">{{ pct(advancedShare) }}</span>
            </v-progress-linear>
            <div class="text-caption text-disabled mt-1">
              {{ (totalMembers - entryTierCount).toLocaleString() }} of
              {{ totalMembers.toLocaleString() }} clients have earned a tier above entry level.
            </div>
          </v-card>
        </v-col>
      </v-row>

      <!-- Close to leveling up -->
      <v-row>
        <v-col cols="12">
          <v-card class="chart-card" variant="flat">
            <div class="chart-card__title">Close to leveling up</div>
            <div class="chart-card__subtitle">
              Clients within reach of their next tier — a nudge here moves them up
            </div>
            <v-data-table
              class="mt-1"
              density="compact"
              :headers="risingHeaders"
              :items="risingClients"
              :items-per-page="10"
              :sort-by="[{ key: 'gapCents', order: 'asc' }]"
            >
              <template #item.name="{ item }">
                <router-link class="client-link" :to="`/clients/${item.id}`">{{ item.name }}</router-link>
              </template>
              <template #item.current="{ item }">
                <v-chip class="text-white" :color="item.currentColor" size="x-small" variant="flat">
                  {{ item.current }}
                </v-chip>
              </template>
              <template #item.spent="{ item }">${{ money(item.spent) }}</template>
              <template #item.gapCents="{ item }">${{ money(item.gapCents) }}</template>
              <template #item.next="{ item }">
                <v-chip class="text-white" :color="item.nextColor" size="x-small" variant="flat">
                  {{ item.next }}
                </v-chip>
              </template>
              <template #item.progress="{ item }">
                <v-progress-linear
                  :color="item.nextColor"
                  height="10"
                  :model-value="item.progress * 100"
                  rounded
                  style="min-width: 90px"
                />
              </template>
              <template #no-data>
                No clients are close to their next tier right now.
              </template>
            </v-data-table>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import VChart from 'vue-echarts'
  import { useTheme } from 'vuetify'
  import { useClientTiers } from '@/composables/useClientTiers'
  import { dedupedFetch } from '@/composables/useRequestDedup'
  import { API_BASE } from '@/config'
  import { formatCompactMoney } from '@/utils/money'

  // --- echarts chrome (same helpers as footfall.vue / conversion.vue) -------
  const theme = useTheme()
  const isDark = computed(() => theme.current.value.dark)
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

  const clientTypeOptions = [
    { title: 'All clients', value: 'all' },
    { title: 'Retail', value: 'retail' },
    { title: 'Wholesale', value: 'wholesale' },
  ]
  const clientType = ref('all')
  const clientTypeLabel = computed(
    () => clientTypeOptions.find(o => o.value === clientType.value)?.title ?? '',
  )

  const { sortedTiers, tierFor } = useClientTiers()
  // Entry tier first, top tier last -- the order the pyramid and the
  // composition bars read in.
  const tierLadder = computed(() => sortedTiers().toReversed())

  // --- data load ----------------------------------------------------------
  // No aggregate endpoint exists for tiers (they're a front-end bucketing
  // over total_spent), so we page the ranked client list and bucket it
  // here. CAP keeps a pathological client count from firing hundreds of
  // requests; because the list is ranked by spend, the truncated tail is
  // all lowest-tier members -- noted in the UI when it happens.
  const PAGE_SIZE = 100
  const CAP = 5000

  const rows = ref([])
  const totalClients = ref(0)
  const loadedCount = ref(0)
  const loading = ref(false)
  const error = ref('')
  const truncated = computed(() => totalClients.value > rows.value.length)

  function pageURL (offset) {
    const params = new URLSearchParams()
    if (clientType.value !== 'all') params.set('client_type', clientType.value)
    params.set('sort_by', 'spending')
    params.set('page_size', String(PAGE_SIZE))
    params.set('page_id', String(offset))
    return `${API_BASE}/clients/by_spending?${params}`
  }

  async function load () {
    loading.value = true
    error.value = ''
    loadedCount.value = 0
    try {
      const first = await dedupedFetch(pageURL(0))
      const collected = first?.clients ?? []
      totalClients.value = first?.total ?? collected.length
      loadedCount.value = collected.length

      const target = Math.min(totalClients.value, CAP)
      while (collected.length < target) {
        const batch = await dedupedFetch(pageURL(collected.length))
        const chunk = batch?.clients ?? []
        if (chunk.length === 0) break
        collected.push(...chunk)
        loadedCount.value = collected.length
      }
      rows.value = collected
    } catch (error_) {
      error.value = error_.message
      rows.value = []
      totalClients.value = 0
    } finally {
      loading.value = false
    }
  }

  onMounted(load)
  watch(clientType, load)

  // --- aggregation ------------------------------------------------------------
  function money (cents) {
    return (Number(cents || 0) / 100).toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  }
  const pct = x => `${(x * 100).toFixed(1)}%`

  const totalMembers = computed(() => rows.value.length)
  const activeMembers = computed(
    () => rows.value.filter(c => (c.invoice_count ?? 0) > 0).length,
  )
  const totalRevenue = computed(
    () => rows.value.reduce((s, c) => s + (c.total_spent ?? 0), 0),
  )
  const totalPoints = computed(
    () => rows.value.reduce((s, c) => s + (c.valid_loyalty_points ?? 0), 0),
  )

  // One bucket per tier, entry-first. Every client lands in exactly one via
  // the same tierFor() the Clients page uses.
  const tierStats = computed(() => {
    const ladder = tierLadder.value
    const buckets = new Map(
      ladder.map(t => [t.id, { id: t.id, label: t.label, color: t.color, minAmount: t.minAmount, members: 0, active: 0, revenue: 0, points: 0 }]),
    )
    for (const c of rows.value) {
      const bucket = buckets.get(tierFor(c.total_spent)?.id)
      if (!bucket) continue
      bucket.members++
      if ((c.invoice_count ?? 0) > 0) bucket.active++
      bucket.revenue += c.total_spent ?? 0
      bucket.points += c.valid_loyalty_points ?? 0
    }
    const members = rows.value.length
    const revenue = totalRevenue.value
    return ladder.map(t => {
      const b = buckets.get(t.id)
      return {
        ...b,
        baseShare: members ? b.members / members : 0,
        revShare: revenue ? b.revenue / revenue : 0,
        avgSpend: b.members ? b.revenue / b.members : 0,
      }
    })
  })
  // Table reads top-tier first; charts/bars read entry-first.
  const tierStatsDesc = computed(() => tierStats.value.toReversed())

  const entryTier = computed(() => tierLadder.value[0])
  const topTier = computed(() => tierLadder.value.at(-1))
  const entryTierCount = computed(
    () => tierStats.value.find(t => t.id === entryTier.value?.id)?.members ?? 0,
  )
  const advancedShare = computed(
    () => (totalMembers.value ? (totalMembers.value - entryTierCount.value) / totalMembers.value : 0),
  )
  const topTierRevShare = computed(
    () => tierStats.value.find(t => t.id === topTier.value?.id)?.revShare ?? 0,
  )

  const kpis = computed(() => [
    { label: 'Members', value: totalMembers.value.toLocaleString(), hint: clientTypeLabel.value },
    { label: 'Active members', value: activeMembers.value.toLocaleString(), hint: '≥ 1 purchase' },
    { label: 'Lifetime revenue', value: formatCompactMoney(totalRevenue.value) },
    { label: 'Points outstanding', value: totalPoints.value.toLocaleString(), hint: 'unredeemed balance' },
    {
      label: 'Avg spend / member',
      value: totalMembers.value ? formatCompactMoney(totalRevenue.value / totalMembers.value) : '—',
    },
    { label: `${topTier.value?.label ?? 'Top'} revenue share`, value: pct(topTierRevShare.value) },
  ])

  // --- charts -----------------------------------------------------------------
  const membersByTierOption = computed(() => ({
    grid: { left: 8, right: 40, top: 10, bottom: 8, containLabel: true },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: p => `${p[0].name}: ${Number(p[0].value).toLocaleString()} members`,
    },
    xAxis: { type: 'value', ...valAxis.value },
    yAxis: { type: 'category', data: tierStats.value.map(t => t.label), ...catAxis.value },
    series: [
      {
        type: 'bar',
        barWidth: '55%',
        data: tierStats.value.map(t => ({
          value: t.members,
          itemStyle: { color: t.color, borderRadius: [0, 4, 4, 0] },
        })),
        label: {
          show: true,
          position: 'right',
          color: textColor.value,
          fontSize: 11,
          formatter: p => Number(p.value).toLocaleString(),
        },
      },
    ],
  }))

  const revenueShareOption = computed(() => ({
    tooltip: {
      trigger: 'item',
      formatter: p => `${p.name}: $${Number(p.value).toLocaleString()} (${p.percent}%)`,
    },
    series: [
      {
        type: 'pie',
        radius: ['42%', '70%'],
        center: ['50%', '48%'],
        data: tierStats.value
          .filter(t => t.revenue > 0)
          .map(t => ({ name: t.label, value: Math.round(t.revenue / 100), itemStyle: { color: t.color } })),
        label: { color: textColor.value, fontSize: 11, formatter: '{b}\n{d}%' },
        labelLine: { length: 8, length2: 6 },
      },
    ],
  }))

  const pointsByTierOption = computed(() => ({
    grid: { left: 8, right: 40, top: 10, bottom: 8, containLabel: true },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: p => `${p[0].name}: ${Number(p[0].value).toLocaleString()} pts`,
    },
    xAxis: { type: 'value', ...valAxis.value },
    yAxis: { type: 'category', data: tierStats.value.map(t => t.label), ...catAxis.value },
    series: [
      {
        type: 'bar',
        barWidth: '55%',
        data: tierStats.value.map(t => ({
          value: t.points,
          itemStyle: { color: t.color, borderRadius: [0, 4, 4, 0] },
        })),
      },
    ],
  }))

  const monthlyJoins = computed(() => {
    const now = new Date()
    const keys = []
    for (let i = 11; i >= 0; i--) {
      const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
      keys.push(`${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`)
    }
    const counts = Object.fromEntries(keys.map(k => [k, 0]))
    for (const c of rows.value) {
      const k = String(c.created_at ?? '').slice(0, 7)
      if (k in counts) counts[k]++
    }
    return { keys, values: keys.map(k => counts[k]) }
  })

  const newMembersOption = computed(() => ({
    grid: { left: 8, right: 16, top: 16, bottom: 8, containLabel: true },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: monthlyJoins.value.keys.map(k => {
        const [y, m] = k.split('-')
        return new Date(Number(y), Number(m) - 1, 1).toLocaleDateString(undefined, { month: 'short', year: '2-digit' })
      }),
      ...catAxis.value,
    },
    yAxis: { type: 'value', minInterval: 1, ...valAxis.value },
    series: [
      {
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: monthlyJoins.value.values,
        lineStyle: { width: 2 },
        areaStyle: { opacity: 0.1 },
      },
    ],
  }))

  // --- close to leveling up -------------------------------------------------
  const risingHeaders = [
    { title: 'Client', key: 'name' },
    { title: 'Current', key: 'current', sortable: false },
    { title: 'Spend', key: 'spent', align: 'end' },
    { title: 'Gap to next', key: 'gapCents', align: 'end' },
    { title: 'Next tier', key: 'next', sortable: false },
    { title: 'Progress', key: 'progress', align: 'end', sortable: false },
  ]

  const risingClients = computed(() => {
    const ladder = tierLadder.value
    const out = []
    for (const c of rows.value) {
      const current = tierFor(c.total_spent)
      const idx = ladder.findIndex(t => t.id === current?.id)
      if (idx === -1 || idx >= ladder.length - 1) continue // already at the top
      const next = ladder[idx + 1]
      const dollars = (c.total_spent ?? 0) / 100
      const gap = next.minAmount - dollars
      const progress = next.minAmount > 0 ? dollars / next.minAmount : 1
      if (gap <= 0 || progress < 0.7) continue
      out.push({
        id: c.id,
        name: c.name,
        current: current.label,
        currentColor: current.color,
        next: next.label,
        nextColor: next.color,
        spent: c.total_spent ?? 0,
        gapCents: Math.round(gap * 100),
        progress,
      })
    }
    return out.toSorted((a, b) => a.gapCents - b.gapCents).slice(0, 25)
  })
</script>

<style scoped>
/* v-alert defaults to flex: 1 1 inside a flex column -- pin notices to
   their content height so they don't stretch the whole page. */
.ly-notice {
  flex: 0 0 auto;
}

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

.tier-dot {
  display: inline-block;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}

.comp-row {
  display: grid;
  grid-template-columns: 84px 1fr auto;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.comp-label {
  font-size: 0.8rem;
  white-space: nowrap;
}

.comp-value {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.7);
  white-space: nowrap;
}

.client-link {
  color: rgb(var(--v-theme-primary));
  text-decoration: none;
}

.client-link:hover {
  text-decoration: underline;
}
</style>
