<template>
  <v-navigation-drawer
    class="nav-drawer"
    density="compact"
    location="left"
    :model-value="props.mobile ? props.modelValue : true"
    :permanent="!props.mobile"
    :rail="props.rail"
    :temporary="props.mobile"
    width="260"
    @update:model-value="props.mobile && $emit('update:modelValue', $event)"
  >
    <v-list density="compact" nav>
      <template v-for="item in navItems" :key="item.title">
        <!-- Group item -->
        <v-list-group
          v-if="item.children"
          :prepend-icon="item.icon"
          :value="item.title"
        >
          <template #activator="{ props: activatorProps }">
            <v-list-item v-bind="activatorProps" :title="item.title" />
          </template>

          <v-list-item
            v-for="child in item.children"
            :key="child.title"
            class="ms-1 me-0"
            :prepend-icon="child.icon"
            :title="child.title"
            :to="child.to"
            :value="child.to"
          >
            <template v-if="child.append === 'status'" #append>
              <span
                :aria-label="`Stock health status is ${healthStatus}`"
                class="nav-status d-flex align-center ga-2"
                role="img"
              >
                <span class="nav-status-dot" :class="`nav-status-dot--${healthStatus}`" />
              </span>
            </template>
          </v-list-item>
        </v-list-group>

        <!-- Single item -->
        <v-list-item
          v-else
          :color="item.color"
          :prepend-icon="item.icon"
          :title="item.title"
          :to="item.to"
          :value="item.title"
        >
          <template v-if="item.append === 'status'" #append>
            <span
              :aria-label="`Stock health status is ${healthStatus}`"
              class="nav-status d-flex align-center ga-2"
              role="img"
            >
              <span class="nav-status-dot" :class="`nav-status-dot--${healthStatus}`" />
            </span>
          </template>
          <template v-else-if="item.append === 'alert-count' && (outOfStockCount > 0 || lowStockCount > 0)" #append>
            <span class="d-flex align-center ga-1">
              <v-badge
                v-if="outOfStockCount > 0"
                color="error"
                :content="outOfStockCount"
                inline
              />
              <v-badge
                v-if="lowStockCount > 0"
                color="warning"
                :content="lowStockCount"
                inline
              />
            </span>
          </template>
        </v-list-item>
      </template>
    </v-list>

    <template #append>
      <v-divider thickness="2" />
      <v-list class="pa-0" density="compact" nav>
        <div class="pa-2">
          <v-list-item
            v-for="item in appendItems"
            :key="item.title"
            :prepend-icon="item.icon"
            :title="item.title"
            :to="item.to"
            :value="item.title"
          />
        </div>
      </v-list>
    </template>
  </v-navigation-drawer>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { useLowStockAlerts } from '@/composables/useLowStockAlerts'
  import { useStockHealth } from '@/composables/useStockHealth'
  import { appendItems, navItems } from '@/data/navItems'

  const props = defineProps({
    modelValue: { type: Boolean, required: true },
    rail: { type: Boolean, default: false },
    mobile: { type: Boolean, default: false },
  })
  defineEmits(['update:modelValue', 'update:rail'])

  // Real counts backing the Alerts nav badges, shown separately: red for
  // out-of-stock, orange for merely low (see pages/alerts.vue, api/low_stock.go).
  const { fetchSummary } = useLowStockAlerts()
  const outOfStockCount = ref(0)
  const lowStockCount = ref(0)

  onMounted(async () => {
    try {
      const summary = await fetchSummary()
      outOfStockCount.value = summary?.out_of_stock_count ?? 0
      lowStockCount.value = summary?.low_stock_count ?? 0
    } catch {
      outOfStockCount.value = 0
      lowStockCount.value = 0
    }
  })

  // Real status backing the Stock Health nav dot -- green/orange/red for
  // healthy/warning/critical (see pages/stock-health.vue, api/stock_health.go).
  const { fetchStockHealth } = useStockHealth()
  const healthStatus = ref('healthy')

  onMounted(async () => {
    try {
      const health = await fetchStockHealth()
      healthStatus.value = health?.status ?? 'healthy'
    } catch {
      healthStatus.value = 'healthy'
    }
  })
</script>

<style scoped>
.nav-drawer {
  --v-list-prepend-gap: 20px;
}

.nav-drawer :deep(.v-list-group__items .v-list-item) {
  padding-inline-start: 16px !important;
}

.nav-status-dot {
  width: 18px;
  height: 18px;
  display: block;
  border-radius: 50%;
  background-color: rgb(var(--v-theme-success));
}
.nav-status-dot--healthy {
  background-color: rgb(var(--v-theme-success));
}
.nav-status-dot--warning {
  background-color: rgb(var(--v-theme-warning));
}
.nav-status-dot--critical {
  background-color: rgb(var(--v-theme-error));
}

.nav-status {
  min-width: 24px;
  justify-content: center;
  margin-inline-end: 8px;
}
</style>
