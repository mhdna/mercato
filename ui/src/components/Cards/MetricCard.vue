<template>
  <v-card class="metric-card pa-0" rounded="lg" variant="outlined" width="100%">
    <v-icon
      class="metric-card-icon"
      color="indigo-lighten-2"
      :icon="props.icon"
      size="26"
    />

    <div class="metric-card-body">
      <div class="text-body-small text-uppercase text-medium-emphasis">
        {{ props.title }}
      </div>

      <!-- Only the very first load (no value yet) shows a spinner. Once a
        number is on screen it stays put through every background refresh. -->
      <v-progress-circular
        v-if="props.loading && props.value == null"
        class="mt-2"
        color="indigo-lighten-2"
        indeterminate
        size="22"
        width="2"
      />
      <template v-else>
        <div class="metric-value text-headline-small font-weight-black mt-1">
          <strong v-if="props.unit === '$' && display !== '—'">{{ props.unit }}</strong>{{ display }}
        </div>

        <div class="metric-delta text-caption mt-1" :class="deltaClass">
          <template v-if="deltaPct != null">
            <v-icon :icon="deltaUp ? 'mdi-arrow-up' : 'mdi-arrow-down'" size="14" />
            {{ deltaText }}
            <span class="text-medium-emphasis ms-1">vs last period</span>
          </template>
          <span v-else class="text-medium-emphasis">no prior period</span>
        </div>

        <div class="metric-prev text-caption text-medium-emphasis">
          Prev: {{ prevDisplay }}
        </div>
      </template>
    </div>
  </v-card>
</template>

<script setup>
  import { computed } from 'vue'
  import { formatMoney } from '@/utils/money'

  const props = defineProps({
    title: {
      type: String,
      required: true,
    },
    icon: {
      type: String,
      required: true,
    },
    unit: {
      type: String,
      required: true,
    },
    // Raw metric value from /dashboard/summary. Money metrics ($ unit)
    // arrive in cents. null renders as an em dash.
    value: {
      type: Number,
      default: null,
    },
    // Same metric for the immediately-preceding period.
    prevValue: {
      type: Number,
      default: null,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  })

  function fmt (v) {
    if (v == null || Number.isNaN(v)) return '—'
    if (props.unit === '$') return formatMoney(v)
    if (props.unit === '%') return `${v.toLocaleString()}%`
    return v.toLocaleString()
  }

  const display = computed(() => fmt(props.value))
  const prevDisplay = computed(() => {
    const v = fmt(props.prevValue)
    return props.unit === '$' && v !== '—' ? `${props.unit}${v}` : v
  })

  const deltaPct = computed(() => {
    if (props.value == null || props.prevValue == null || props.prevValue === 0) return null
    return ((props.value - props.prevValue) / Math.abs(props.prevValue)) * 100
  })
  const deltaUp = computed(() => (deltaPct.value ?? 0) >= 0)
  const deltaClass = computed(() => {
    if (deltaPct.value == null) return ''
    return deltaUp.value ? 'text-success' : 'text-error'
  })
  const deltaText = computed(() => {
    if (deltaPct.value == null) return ''
    const sign = deltaPct.value >= 0 ? '+' : '−'
    return `${sign}${Math.abs(deltaPct.value).toFixed(1)}%`
  })
</script>

<style scoped>
  .metric-card {
    position: relative;
  }

  .metric-card-icon {
    position: absolute;
    left: 10px;
    top: 10px;
  }

  .metric-card-body {
    padding: 12px 10px;
    text-align: center;
  }

  /* Fixed-width digits so the number doesn't jitter sideways between
     filter/period selections. */
  .metric-value,
  .metric-prev {
    font-variant-numeric: tabular-nums;
  }
</style>
