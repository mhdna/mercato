<template>
  <v-menu v-model="menu" :close-on-content-click="false" location="bottom end">
    <template #activator="{ props }">
      <v-card
        v-bind="props"
        class="d-flex align-center py-1 px-2"
        color="transparent"
        flat
        rounded="xl"
        style="cursor: pointer"
      >
        <div class="me-2 text-body-2">
          {{ statusText }}
        </div>
        <v-icon :color="wsIconColor" icon="mdi-cloud" />
      </v-card>
    </template>

    <v-card max-width="400" min-width="320">
      <v-card-title class="text-subtitle-1">Branch sync</v-card-title>
      <v-divider />

      <v-list v-if="recentInvoices.length > 0" density="compact" lines="two">
        <v-list-item
          v-for="invoice in recentInvoices"
          :key="invoice.id"
          :prepend-icon="invoice.kind === 'return' ? 'mdi-transfer' : 'mdi-invoice'"
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
  import { useBranches } from '@/composables/useBranches'
  import { useBranchInvoices } from '@/composables/useBranchInvoices'

  const { branches, fetchBranches } = useBranches()
  const { listRecentBranchInvoices } = useBranchInvoices()
  const { status: wsStatus, ensureConnected, onMessage } = useAdminSocket()

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

  const wsIconColor = computed(() => (wsStatus.value === 'open' ? 'success' : undefined))

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

  function invoiceTitle (invoice) {
    const kind = invoice.kind === 'return' ? 'Return' : 'Sale'
    return `${kind} ${invoice.branch_invoice_code}`
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
    })
  })

  onUnmounted(() => {
    clearInterval(refreshTimer)
    clearInterval(clockTimer)
    unsubscribe?.()
  })
</script>
