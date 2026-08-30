<template>
  <div class="d-flex flex-wrap ga-3 mb-n3">
    <!-- Targets -->
    <v-card class="pa-4" flat style="flex: 1 1 320px; min-width: 0">
      <div class="d-flex justify-space-between align-center mb-3">
        <span class="d-flex align-center text-subtitle-2 text-uppercase text-medium-emphasis">
          <v-icon class="me-2" color="indigo-lighten-2" icon="mdi-target" size="20" />
          Targets
        </span>
        <span class="text-subtitle-1 font-weight-bold">{{ overallPct }}</span>
      </div>

      <v-skeleton-loader v-if="!props.summary" type="text@3" />

      <div v-else-if="targets.length === 0" class="text-body-2 text-medium-emphasis">
        No active targets
      </div>

      <div v-for="(t, i) in targets" v-else :key="i" class="mb-3">
        <div class="d-flex justify-space-between align-center">
          <span class="text-body-2 text-truncate">{{ t.branch_name }}</span>
          <span class="text-caption text-medium-emphasis flex-shrink-0 ps-2">
            {{ fmt(t.revenue) }} / {{ fmt(t.target_amount) }} · {{ t.pct }}%
          </span>
        </div>
        <v-progress-linear
          bg-opacity="0.15"
          class="mt-1"
          :color="barColor(t.color)"
          height="6"
          :model-value="Math.min(t.pct, 100)"
          rounded
        />
      </div>
    </v-card>

    <!-- Due Loans -->
    <v-card class="pa-4" flat style="flex: 1 1 320px; min-width: 0">
      <div class="d-flex justify-space-between align-center mb-3">
        <span class="d-flex align-center text-subtitle-2 text-uppercase text-medium-emphasis">
          <v-icon class="me-2" color="indigo-lighten-2" icon="mdi-hand-coin" size="20" />
          Due Loans
        </span>
        <span class="text-subtitle-1 font-weight-bold">{{ props.summary ? fmt(props.summary.due_loans) : '—' }}</span>
      </div>

      <v-skeleton-loader v-if="!props.summary" type="text@3" />

      <div v-else-if="loans.length === 0" class="text-body-2 text-medium-emphasis">
        No outstanding loans
      </div>

      <div
        v-for="(l, i) in loans"
        v-else
        :key="i"
        class="d-flex justify-space-between align-center py-1"
      >
        <div class="d-flex align-center text-truncate">
          <span class="text-body-2 text-truncate">{{ l.description || 'Loan' }}</span>
          <v-chip
            class="ms-2 flex-shrink-0"
            :color="l.origin === 'branch_loan' ? 'primary' : 'secondary'"
            size="x-small"
          >
            {{ l.origin === 'branch_loan' ? (l.branch_name || 'Branch') : 'Central' }}
          </v-chip>
        </div>
        <span class="text-body-2 font-weight-medium flex-shrink-0 ps-2">{{ fmt(l.outstanding) }}</span>
      </div>
    </v-card>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { formatCompactMoney as fmt } from '@/utils/money'

  const props = defineProps({
    // The whole /dashboard/summary payload (or null while first loading).
    summary: {
      type: Object,
      default: null,
    },
  })

  const targets = computed(() => props.summary?.targets ?? [])
  const loans = computed(() => props.summary?.due_loans_top ?? [])

  const overallPct = computed(() => {
    const p = props.summary?.targets_pct
    return typeof p === 'number' ? `${p}%` : '—'
  })

  function barColor (color) {
    return color && color.trim() ? color : 'primary'
  }
</script>
