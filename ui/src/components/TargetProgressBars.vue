<template>
  <div v-if="sorted.length > 0" class="target-stack">
    <div v-for="target in sorted" :key="target.id" class="target-row">
      <div class="target-track" :style="{ width: target.widthPct + '%' }">
        <div class="target-fill" :style="{ width: target.progressPct + '%', background: target.color || undefined }" />
      </div>
      <div class="target-meta">
        <span class="target-range">{{ formatDate(target.date_from) }} – {{ formatDate(target.date_to) }}</span>
        <span class="target-amounts">{{ formatMoney(target.achieved) }} / {{ formatMoney(target.target_amount) }}
          ({{ Math.round(target.progressPct) }}%)</span>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed } from 'vue'

  // `targets` amounts (target_amount, achieved) are in cents, matching
  // branch_invoices.grand_total's convention (see 000025_create_branch_invoices).
  const props = defineProps({
    targets: {
      type: Array,
      default: () => [],
    },
  })

  const sorted = computed(() => {
    const list = props.targets.toSorted((a, b) => a.target_amount - b.target_amount)
    const maxAmount = list.length > 0 ? list.at(-1).target_amount : 1
    return list.map(target => ({
      ...target,
      // Lower targets get a wider track (up to 100%); higher targets are
      // progressively narrower -- a "funnel" of nested goals, easiest/
      // nearest at the top and full width, hardest/farthest below and thin.
      widthPct: maxAmount > 0 ? (target.target_amount / maxAmount) * 100 : 100,
      progressPct: target.target_amount > 0
        ? Math.min(100, (target.achieved / target.target_amount) * 100)
        : 0,
    }))
  })

  function formatMoney (cents) {
    return `$${(Number(cents ?? 0) / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  function formatDate (value) {
    return new Date(value).toLocaleDateString()
  }
</script>

<style scoped>
.target-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.target-row {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.target-track {
  height: 10px;
  border-radius: 5px;
  background: rgba(var(--v-theme-on-surface), 0.08);
  overflow: hidden;
  transition: width 0.2s ease;
}

.target-fill {
  height: 100%;
  border-radius: 5px;
  background: rgb(var(--v-theme-primary));
  transition: width 0.2s ease;
}

.target-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.7);
}
</style>
