<template>
  <v-menu v-model="menu" :close-on-content-click="false" location="bottom center" :offset="8">
    <template #activator="{ props }">
      <v-card
        v-bind="props"
        class="d-flex align-center py-1 px-2 me-2"
        rounded="xl"
        style="cursor: pointer"
        variant="tonal"
      >
        <v-icon
          class="me-2"
          :color="currentActivity ? currentActivity.color : wsIconColor"
          icon="mdi-cloud"
        />
        <div class="text-body-2 ticker">
          <Transition mode="out-in" name="ticker">
            <span :key="displayKey">
              <v-icon
                v-if="currentActivity && currentActivity.trendIcon"
                class="me-1"
                :color="currentActivity.color"
                :icon="currentActivity.trendIcon"
                size="16"
              />
              {{ displayText }}
            </span>
          </Transition>
        </div>
      </v-card>
    </template>

    <v-card max-width="400" min-width="320">
      <v-card-title class="text-subtitle-1">Branch sync</v-card-title>
      <v-divider />

      <v-card-text v-if="serverDown" class="text-medium-emphasis">
        Can't reach the server, we'll be back soon.
      </v-card-text>
      <v-list v-else-if="recentInvoices.length > 0" density="compact" lines="two">
        <v-list-item
          v-for="invoice in recentInvoices"
          :key="invoice.id"
          :prepend-icon="invoiceIcon(invoice)"
          :subtitle="`${branchName(invoice.branch_id)} — ${relativeTime(invoice.received_at)}`"
          :title="invoiceTitle(invoice)"
        />
      </v-list>
      <v-card-text v-else-if="loaded" class="text-medium-emphasis">
        No branch data has synced yet.
      </v-card-text>
      <v-card-text v-else class="text-medium-emphasis">
        Loading...
      </v-card-text>
    </v-card>
  </v-menu>
</template>

<script setup>
  import { computed, onMounted, onUnmounted, ref } from 'vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { describeBranchActivity, isBranchActivityMessage } from '@/composables/useBranchActivityMessage'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchInvoices } from '@/composables/useBranchInvoices'
  import { useSettingsStore } from '@/stores/settings'

  const { branches, fetchBranches } = useBranches()
  const { listRecentBranchInvoices } = useBranchInvoices()
  const { status: wsStatus, ensureConnected, onMessage } = useAdminSocket()
  const settingsStore = useSettingsStore()

  const menu = ref(false)
  const loaded = ref(false)
  const recentInvoices = ref([])
  // Ticks once a minute purely to force relativeTime()/statusText to
  // re-evaluate -- received_at/last_seen_at themselves only change on refresh.
  const now = ref(Date.now())

  // Live push is purely an optimization on top of the poll below -- if the
  // socket never connects or drops, the 30s poll still keeps this card
  // correct, just slower. wsStatus only drives the icon color.
  let unsubscribe = null

  let refreshTimer = null
  let clockTimer = null

  // No live socket means we can't confirm branch data is flowing, so the
  // card reports "not syncing" rather than a stale "synced ... ago".
  const serverDown = computed(() => wsStatus.value !== 'open')
  const wsIconColor = computed(() => (serverDown.value ? undefined : 'success'))

  // 'appbar' mode: branch activity (sales/returns/expenses) takes over the
  // card's single line for a few seconds at a time, one message at a time
  // -- queued rather than shown all at once, so a burst of simultaneous
  // branch activity doesn't get lost, just delayed. Once the queue drains,
  // the card falls back to the normal "synced ... ago" text. In
  // 'notification' mode (the default), BranchActivityToast owns this
  // instead and this queue is never fed -- see the onMessage handler below.
  const activityQueue = ref([])
  const currentActivity = ref(null)
  // Bumped on every cycle (including the final flip back to statusText) so
  // the ticker <Transition> below always sees a fresh :key and replays its
  // slide animation -- without this, statusText's own passive updates
  // (relative time ticking every 30s) would either wrongly replay the
  // transition or, if keyed by text, fail to replay on a repeated amount.
  const activitySeq = ref(0)
  let activityTimer = null

  const displayText = computed(() => (currentActivity.value ? currentActivity.value.text : statusText.value))
  const displayKey = computed(() => (currentActivity.value ? `activity-${activitySeq.value}` : 'status'))

  function advanceActivity () {
    activitySeq.value++
    if (activityQueue.value.length === 0) {
      currentActivity.value = null
      activityTimer = null
      return
    }
    currentActivity.value = activityQueue.value.shift()
    activityTimer = setTimeout(advanceActivity, settingsStore.activityMessageSeconds * 1000)
  }

  function pushActivity (message) {
    const described = describeBranchActivity(message, branchName)
    if (!described) return
    activityQueue.value.push({
      text: described.text,
      color: described.color,
      trendIcon: described.trendIcon,
    })
    if (!activityTimer) {
      advanceActivity()
    }
  }

  async function refresh () {
    try {
      const [invoices] = await Promise.all([
        listRecentBranchInvoices(5),
        fetchBranches(true),
      ])
      recentInvoices.value = invoices
    } catch {
      // Sync status is informational -- a failed refresh just keeps
      // showing the last known state rather than surfacing an error.
    } finally {
      loaded.value = true
      now.value = Date.now()
    }
  }

  function branchName (branchId) {
    return branches.value.find(b => b.id === branchId)?.name ?? `Branch #${branchId}`
  }

  function invoiceIcon (invoice) {
    if (invoice.kind === 'exchange') return 'mdi-swap-horizontal'
    if (invoice.kind === 'return') return 'mdi-transfer'
    return 'mdi-invoice'
  }

  function invoiceKindLabel (invoice) {
    if (invoice.kind === 'exchange') return 'Exchange'
    if (invoice.kind === 'return') return 'Return'
    return 'Sale'
  }

  function invoiceTitle (invoice) {
    return `${invoiceKindLabel(invoice)} ${invoice.branch_invoice_code}`
  }

  function relativeTime (isoString) {
    const then = new Date(isoString).getTime()
    const diffSeconds = Math.max(0, Math.round((now.value - then) / 1000))

    if (diffSeconds < 45) return 'just now'
    const diffMinutes = Math.round(diffSeconds / 60)
    if (diffMinutes < 60) return `${diffMinutes} min${diffMinutes === 1 ? '' : 's'} ago`
    const diffHours = Math.round(diffMinutes / 60)
    if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`
    const diffDays = Math.round(diffHours / 24)
    return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`
  }

  const statusText = computed(() => {
    if (serverDown.value) return 'not syncing'
    if (!loaded.value) return 'syncing...'
    const latest = recentInvoices.value[0]
    if (!latest) return 'no sync yet'
    return `synced ${relativeTime(latest.received_at)}`
  })

  onMounted(() => {
    refresh()
    refreshTimer = setInterval(refresh, 30_000)
    clockTimer = setInterval(() => {
      now.value = Date.now()
    }, 30_000)
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (message.type === 'branch_invoice_created') {
        refresh()
      }
      if (settingsStore.activityDisplayMode === 'appbar' && isBranchActivityMessage(message)) {
        pushActivity(message)
      }
    })
  })

  onUnmounted(() => {
    clearInterval(refreshTimer)
    clearInterval(clockTimer)
    clearTimeout(activityTimer)
    unsubscribe?.()
  })
</script>

<style scoped>
/* Stock-ticker-style vertical roll: the outgoing line slides up and out
   while the incoming one slides up and in from below, like a flip/odometer
   readout rather than a plain crossfade. */
.ticker {
  position: relative;
  overflow: hidden;
  white-space: nowrap;
}

.ticker-enter-active,
.ticker-leave-active {
  transition: transform 0.35s ease, opacity 0.35s ease;
}

.ticker-enter-from {
  transform: translateY(100%);
  opacity: 0;
}

.ticker-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}
</style>
