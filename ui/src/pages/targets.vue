<template>
  <div class="targets d-flex flex-column fill-height" style="overflow-y: auto">
    <!-- Header / controls -->
    <div class="d-flex align-center flex-wrap ga-3 pa-4 pb-2">
      <v-icon icon="mdi-target" />
      <div class="text-h6 font-weight-medium">Targets</div>
      <v-spacer />
      <v-select
        v-model="selectedPeriod"
        density="compact"
        hide-details
        :items="periodOptions"
        style="max-width: 190px"
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
      class="tg-notice mx-4 mb-2"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ error }}
    </v-alert>

    <div v-if="loading" class="d-flex justify-center pa-12">
      <v-progress-circular color="primary" indeterminate />
    </div>

    <v-alert v-else-if="allTargets.length === 0" class="tg-notice mx-4" type="info" variant="tonal">
      No targets set on any branch yet.
    </v-alert>

    <v-container v-else class="pa-4 pt-2" fluid>
      <!-- Scoreboard: click a card to filter the table below to that bucket -->
      <v-row dense>
        <v-col v-for="kpi in kpis" :key="kpi.bucket" cols="6" md="3">
          <v-card
            class="pa-3 d-flex flex-column justify-center fill-height kpi"
            :class="{ 'kpi--active': bucket === kpi.bucket }"
            :color="kpi.color"
            variant="tonal"
            @click="toggleBucket(kpi.bucket)"
          >
            <div class="text-h6 font-weight-medium">{{ kpi.value }}</div>
            <div class="text-caption text-medium-emphasis mt-1">{{ kpi.label }}</div>
          </v-card>
        </v-col>
      </v-row>

      <!-- The targets behind the selected card -->
      <v-row>
        <v-col cols="12">
          <v-card class="chart-card" variant="flat">
            <div class="d-flex align-center chart-card__title">
              <span>{{ detailTitle }}</span>
              <v-spacer />
              <v-btn
                v-if="bucket !== 'all'"
                density="compact"
                size="small"
                variant="text"
                @click="bucket = 'all'"
              >
                Clear filter
              </v-btn>
            </div>

            <v-alert
              v-if="detailRows.length === 0"
              class="tg-notice mt-2"
              density="compact"
              type="info"
              variant="tonal"
            >
              Nothing in this bucket for the selected period.
            </v-alert>

            <div v-else class="table-scroll">
              <v-table density="compact">
                <thead>
                  <tr>
                    <th class="text-left">Branch</th>
                    <th class="text-left">Period</th>
                    <th class="text-right">Achieved</th>
                    <th class="text-right">Target</th>
                    <th class="text-right">Progress</th>
                    <th class="text-left">Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="t in detailRows" :key="t.id">
                    <td>{{ t.branchName }}</td>
                    <td>{{ formatDate(t.date_from) }} – {{ formatDate(t.date_to) }}</td>
                    <td class="text-right">{{ formatMoney(t.achieved) }}</td>
                    <td class="text-right">{{ formatMoney(t.target_amount) }}</td>
                    <td class="text-right">{{ Math.round(t.progressPct) }}%</td>
                    <td>
                      <v-chip :color="STATUS_META[t.status].color" size="small" variant="tonal">
                        {{ STATUS_META[t.status].label }}
                      </v-chip>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref, watch } from 'vue'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchTargets } from '@/composables/useBranchTargets'

  const { branches, fetchBranches } = useBranches()
  const { listBranchTargets } = useBranchTargets()

  const loading = ref(false)
  const error = ref('')
  // One flat list across every branch, each row carrying its branch name and
  // a derived status. The API already returns `achieved` per target, so met /
  // missed is purely client-side arithmetic -- no extra endpoint.
  const allTargets = ref([])

  const selectedBranchIds = ref([]) // empty = all
  const selectedPeriod = ref('all') // 'all' or a 'YYYY-MM' key
  const bucket = ref('all') // 'all' | 'met' | 'missed' | 'in_progress'

  const STATUS_META = {
    met: { label: 'Met', color: 'success' },
    missed: { label: 'Missed', color: 'error' },
    in_progress: { label: 'In progress', color: 'info' },
  }

  const todayKey = new Date().toISOString().slice(0, 10)

  function monthKey (value) {
    return value.slice(0, 7)
  }
  function monthLabel (key) {
    const [y, m] = key.split('-')
    return new Date(Number(y), Number(m) - 1, 1)
      .toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
  }
  function formatDate (value) {
    return new Date(value).toLocaleDateString()
  }
  function formatMoney (cents) {
    return `$${(Number(cents ?? 0) / 100).toLocaleString(undefined, { maximumFractionDigits: 0 })}`
  }

  // A target's period is judged by its end date: met once achieved reaches
  // the goal, missed if the window has closed short, otherwise still running.
  function statusOf (t) {
    if (t.achieved >= t.target_amount) return 'met'
    if (t.date_to.slice(0, 10) < todayKey) return 'missed'
    return 'in_progress'
  }

  async function load () {
    loading.value = true
    error.value = ''
    try {
      await fetchBranches()
      const perBranch = await Promise.all(
        branches.value.map(async branch => {
          const targets = await listBranchTargets(branch.id)
          return targets.map(t => ({
            ...t,
            branchId: branch.id,
            branchName: branch.name,
            status: statusOf(t),
            progressPct: t.target_amount > 0
              ? Math.min(100, (t.achieved / t.target_amount) * 100)
              : 0,
          }))
        }),
      )
      allTargets.value = perBranch.flat()
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  }
  onMounted(load)

  const branchOptions = computed(() =>
    [...branches.value]
      .toSorted((a, b) => a.name.localeCompare(b.name))
      .map(b => ({ title: b.name, value: b.id })),
  )

  // Every calendar month that has at least one target, newest first, plus an
  // "all periods" catch-all -- this is the "see previous months" control.
  const periodOptions = computed(() => {
    const keys = [...new Set(allTargets.value.map(t => monthKey(t.date_to)))].toSorted((a, b) => b.localeCompare(a))
    return [
      { title: 'All periods', value: 'all' },
      ...keys.map(k => ({ title: monthLabel(k), value: k })),
    ]
  })

  // Targets in scope for the scoreboard: matching the branch filter and,
  // unless "all periods", ending in the selected month.
  const scopedTargets = computed(() => {
    const sel = selectedBranchIds.value
    return allTargets.value.filter(t => {
      if (sel.length > 0 && !sel.includes(t.branchId)) return false
      if (selectedPeriod.value !== 'all' && monthKey(t.date_to) !== selectedPeriod.value) return false
      return true
    })
  })

  const counts = computed(() => {
    const out = { met: 0, missed: 0, in_progress: 0 }
    for (const t of scopedTargets.value) out[t.status] += 1
    return out
  })

  const kpis = computed(() => {
    const { met, missed, in_progress: inProgress } = counts.value
    const decided = met + missed
    return [
      { bucket: 'all', label: 'Total targets', value: scopedTargets.value.length, color: undefined },
      { bucket: 'met', label: 'Targets met', value: met, color: 'success' },
      { bucket: 'missed', label: 'Targets missed', value: missed, color: 'error' },
      {
        bucket: 'in_progress',
        label: 'In progress',
        value: inProgress,
        color: 'info',
        // hit rate shown in the label when there's something decided
      },
    ].map(k => k.bucket === 'in_progress' && decided > 0
      ? { ...k, label: `In progress · ${Math.round((met / decided) * 100)}% hit rate` }
      : k)
  })

  function toggleBucket (b) {
    bucket.value = bucket.value === b ? 'all' : b
  }

  const detailRows = computed(() => {
    const rows = bucket.value === 'all'
      ? scopedTargets.value
      : scopedTargets.value.filter(t => t.status === bucket.value)
    // Newest period first, then by branch.
    return [...rows].toSorted((a, b) =>
      b.date_to.localeCompare(a.date_to) || a.branchName.localeCompare(b.branchName),
    )
  })

  const detailTitle = computed(() => {
    const scope = selectedPeriod.value === 'all' ? 'all periods' : monthLabel(selectedPeriod.value)
    const noun = bucket.value === 'all' ? 'targets' : `${STATUS_META[bucket.value].label.toLowerCase()} targets`
    return `${noun} — ${scope}`
  })

  // If the chosen period drops out of the list (shouldn't, but branch filter
  // changes the data), fall back to "all periods".
  watch(periodOptions, opts => {
    if (!opts.some(o => o.value === selectedPeriod.value)) selectedPeriod.value = 'all'
  })
</script>

<style scoped>
/* v-alert defaults to flex: 1 1 inside a flex column, which stretched the
   empty / error notice to the whole page -- pin it to its content height. */
.tg-notice {
  flex: 0 0 auto;
}
.kpi {
  cursor: pointer;
  transition: outline-color 0.15s ease;
  outline: 2px solid transparent;
  outline-offset: -2px;
}
.kpi--active {
  outline-color: currentColor;
}
.chart-card {
  padding: 12px 14px;
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
.table-scroll {
  overflow-x: auto;
  padding-bottom: 8px;
}
</style>
