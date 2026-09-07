<template>
  <v-menu v-model="menu" :close-on-content-click="false" location="bottom center" :offset="8">
    <template #activator="{ props }">
      <v-card
        v-bind="props"
        class="d-flex align-center me-2"
        :class="mobile ? 'icon-pill' : 'py-1 px-2'"
        rounded="xl"
        style="cursor: pointer"
        variant="tonal"
      >
        <v-icon
          :class="{ 'me-2': !mobile }"
          :color="activeItems.length > 0 ? activeItems[0].color : wsIconColor"
          icon="mdi-cloud"
          :size="mobile ? 20 : undefined"
        />
        <div v-if="!mobile" class="text-body-2 ticker">
          <TransitionGroup class="d-flex align-center reel" name="ticker" tag="div">
            <span
              v-for="item in activeItems"
              :key="`activity-${item.id}`"
              class="activity-item"
            >
              <v-icon
                v-if="item.trendIcon"
                class="me-1"
                :color="item.color"
                :icon="item.trendIcon"
                size="14"
              />
              <span class="activity-item__text">{{ item.text }}</span>
            </span>
            <span v-if="activeItems.length === 0" key="status">{{ statusText }}</span>
          </TransitionGroup>
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
  import { useDisplay } from 'vuetify'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { describeBranchActivity, isBranchActivityMessage } from '@/composables/useBranchActivityMessage'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchInvoices } from '@/composables/useBranchInvoices'
  import { useSettingsStore } from '@/stores/settings'

  const { branches, fetchBranches } = useBranches()
  const { listRecentBranchInvoices } = useBranchInvoices()
  const { status: wsStatus, ensureConnected, onMessage } = useAdminSocket()
  const settingsStore = useSettingsStore()
  const { mobile } = useDisplay()

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

  // 'appbar' mode: branch activity (sales/returns/expenses) rides on the
  // card's line. It's a sliding window of the most recent MAX_VISIBLE
  // messages, newest on the right -- a fresh message shows immediately and
  // bumps the oldest one out rather than waiting in a queue, so a burst
  // scrolls straight through. Each message also carries its own
  // activityMessageSeconds lifetime, so once activity stops the window
  // drains over that long and the card falls back to "synced ... ago". In
  // 'notification' mode (the default) BranchActivityToast owns this instead
  // and activeItems is never fed -- see the onMessage handler.
  const MAX_VISIBLE = 3
  const activeItems = ref([]) // [{ id, text, color, trendIcon }], oldest first
  const itemTimers = new Map()
  let itemSeq = 0

  function dropItem (id) {
    clearTimeout(itemTimers.get(id))
    itemTimers.delete(id)
    const i = activeItems.value.findIndex(it => it.id === id)
    if (i !== -1) activeItems.value.splice(i, 1)
  }

  function pushActivity (message) {
    const described = describeBranchActivity(message, branchName)
    if (!described) return
    const item = {
      id: ++itemSeq,
      text: described.shortText ?? described.text,
      color: described.color,
      trendIcon: described.trendIcon,
    }
    activeItems.value.push(item)
    itemTimers.set(item.id, setTimeout(() => dropItem(item.id), settingsStore.activityMessageSeconds * 1000))
    // Newer message arrived while the window was full -- evict the oldest
    // now, no waiting.
    while (activeItems.value.length > MAX_VISIBLE) {
      dropItem(activeItems.value[0].id)
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
    for (const timer of itemTimers.values()) clearTimeout(timer)
    unsubscribe?.()
  })
</script>

<style scoped>
/* On mobile the pill shrinks to just its icon -- fix it to a square that
   matches ConnectedBranchesCard's icon-only pill. */
.icon-pill {
  width: 34px;
  height: 34px;
  justify-content: center;
}

.ticker {
  position: relative;
  overflow: hidden;
  white-space: nowrap;
  max-width: min(44vw, 520px);
}

.reel {
  gap: 6px;
}

/* Each message is a self-contained pill -- no pseudo-element separators
   between siblings, so removing one only reflows by that pill's own width
   and the survivors' FLIP shift is a single clean amount. */
.activity-item {
  display: inline-flex;
  align-items: center;
  max-width: 220px;
  padding: 1px 8px;
  border-radius: 6px;
  background: rgba(var(--v-theme-on-surface), 0.06);
}

.activity-item__text {
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Newest fades in from the right, the oldest (always the one that leaves)
   fades out and collapses to zero width in place while the survivors glide
   over via the FLIP-driven ticker-move. Animating the leaver's own
   max-width/padding/margin -- rather than yanking it out with
   position:absolute -- is what keeps the exit smooth. */
.ticker-enter-active,
.ticker-leave-active {
  transition:
    opacity 0.24s ease,
    transform 0.24s ease,
    max-width 0.24s ease,
    padding 0.24s ease,
    margin 0.24s ease;
}

.ticker-enter-from {
  transform: translateX(10px);
  opacity: 0;
}

.ticker-leave-to {
  max-width: 0;
  padding-left: 0;
  padding-right: 0;
  margin-left: -6px;
  opacity: 0;
}

.ticker-move {
  transition: transform 0.24s cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
