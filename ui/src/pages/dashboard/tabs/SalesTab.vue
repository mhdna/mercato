<template>
  <div class="sales-tab pa-4" style="min-height: 0; overflow-y: auto">
    <v-alert
      v-if="error"
      class="mb-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ error }}
    </v-alert>
    <div class="summary-grid">
      <MetricCard
        icon="mdi-cart-check"
        :loading="loading"
        :prev-value="prevMetric('items_sold')"
        title="Items Sold"
        unit=""
        :value="metric('items_sold')"
      />
      <MetricCard
        icon="mdi-card-account-details"
        :loading="loading"
        :prev-value="prevMetric('new_clients')"
        title="New Clients"
        unit=""
        :value="metric('new_clients')"
      />
      <MetricCard
        icon="mdi-invoice"
        :loading="loading"
        :prev-value="prevMetric('invoices')"
        title="Invoices"
        unit=""
        :value="metric('invoices')"
      />
      <MetricCard
        icon="mdi-hand-coin"
        :loading="loading"
        :prev-value="prevMetric('due_loans')"
        title="Due Loans"
        unit="$"
        :value="metric('due_loans')"
      />
      <MetricCard
        icon="mdi-package-variant"
        :loading="loading"
        :prev-value="prevMetric('purchases')"
        title="Purchases"
        unit="$"
        :value="metric('purchases')"
      />
      <MetricCard
        icon="mdi-target"
        :loading="loading"
        :prev-value="prevMetric('targets_pct')"
        title="Targets"
        unit="%"
        :value="metric('targets_pct')"
      />
      <MetricCard
        icon="mdi-receipt-text-check"
        :loading="loading"
        :prev-value="prevAvgInvoice"
        title="Avg Invoice"
        unit="$"
        :value="avgInvoice"
      />
      <MetricCard
        icon="mdi-account-cash"
        :loading="loading"
        :prev-value="prevRevPerClient"
        title="Revenue / Client"
        unit="$"
        :value="revPerClient"
      />
    </div>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import MetricCard from '@/components/Cards/MetricCard.vue'
  import { useDashboardSummary } from '@/composables/useDashboardSummary'

  const { summary, prevSummary, loading, error } = useDashboardSummary()

  function metric (key) {
    const value = summary.value?.[key]
    return typeof value === 'number' ? value : null
  }

  function prevMetric (key) {
    const value = prevSummary.value?.[key]
    return typeof value === 'number' ? value : null
  }

  function perUnit (total, count) {
    if (total == null || count == null || count === 0) return null
    return Math.round(total / count)
  }

  const avgInvoice = computed(() => perUnit(metric('revenue'), metric('invoices')))
  const prevAvgInvoice = computed(() => perUnit(prevMetric('revenue'), prevMetric('invoices')))
  const revPerClient = computed(() => perUnit(metric('revenue'), metric('new_clients')))
  const prevRevPerClient = computed(() => perUnit(prevMetric('revenue'), prevMetric('new_clients')))
</script>

<style scoped>
.summary-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 8px;
}
</style>
