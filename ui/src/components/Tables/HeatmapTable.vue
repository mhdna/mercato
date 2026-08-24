<template>
  <div class="heatmap-wrap">
    <table class="heatmap-table">
      <thead>
        <tr>
          <th class="sticky-col day-col" />
          <th v-for="month in months" :key="month">{{ month }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in 31" :key="d">
          <td class="sticky-col day-col">{{ d }}</td>
          <td
            v-for="(month, mi) in months"
            :key="month"
            :style="d <= daysInMonth[mi] ? cellStyle(get(mi, d)) : {}"
          >
            {{ d <= daysInMonth[mi] ? format(get(mi, d)) : "" }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
  import { computed, ref, watch } from 'vue'
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
  const isLeapYear = currentYear % 4 === 0 && (currentYear % 100 !== 0 || currentYear % 400 === 0)
  const daysInMonth = [31, isLeapYear ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]

  const { getDailyIncome } = useDailyIncome()
  const { selectedBranchId } = useIncomeFilters()

  // data[monthIndex][day] = number, cents from the API converted to whole
  // currency units for display, same shape the heatmap always rendered.
  const data = ref(months.map((_, mi) => Array.from({ length: daysInMonth[mi] }, () => 0)))

  async function load () {
    const days = await getDailyIncome(currentYear, selectedBranchId.value).catch(() => [])
    const next = months.map((_, mi) => Array.from({ length: daysInMonth[mi] }, () => 0))
    for (const entry of days) {
      const d = new Date(entry.day)
      next[d.getUTCMonth()][d.getUTCDate() - 1] = entry.total / 100
    }
    data.value = next
  }

  watch(selectedBranchId, load, { immediate: true })

  const get = (mi, day) => data.value[mi][day - 1] ?? 0

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
  min-width: 60px;
  font-weight: 500;
}
</style>
