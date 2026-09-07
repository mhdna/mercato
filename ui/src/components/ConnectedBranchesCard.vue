<template>
  <v-menu v-model="menu" :close-on-content-click="false" location="bottom center" :offset="8">
    <template #activator="{ props }">
      <v-card
        v-bind="props"
        class="d-flex align-center"
        :class="mobile ? 'icon-pill' : 'py-1 px-2'"
        rounded="xl"
        style="cursor: pointer"
        variant="tonal"
      >
        <v-icon
          class="status-dot"
          :class="{ 'status-dot--outlined': serverDown }"
          :color="statusColor"
          icon="mdi-circle"
          size="16"
        />
        <span v-if="!mobile" class="ms-1 text-body-2">
          <template v-if="serverDown">Server is down, we'll be back soon.</template>
          <template v-else>
            {{ connectedBranchIds.length }} {{ connectedBranchIds.length === 1 ? 'device' : 'devices' }} online
          </template>
        </span>
      </v-card>
    </template>

    <v-card max-width="320" min-width="240">
      <v-card-title class="text-subtitle-1">Connected branches</v-card-title>
      <v-divider />

      <v-card-text v-if="serverDown" class="text-medium-emphasis">
        Server is down, we'll be back soon.
      </v-card-text>
      <v-list v-else-if="connectedBranchIds.length > 0" density="compact">
        <v-list-item v-for="id in connectedBranchIds" :key="id" :title="branchName(id)">
          <template #prepend>
            <v-icon color="success" icon="mdi-circle" size="10" />
          </template>
        </v-list-item>
      </v-list>
      <v-card-text v-else class="text-medium-emphasis">
        No branches connected.
      </v-card-text>
    </v-card>
  </v-menu>
</template>

<script setup>
  import { computed, onMounted, onUnmounted, ref } from 'vue'
  import { useDisplay } from 'vuetify'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { useBranches } from '@/composables/useBranches'

  const { branches, fetchBranches, fetchConnectedBranches } = useBranches()
  const { status, ensureConnected, onMessage } = useAdminSocket()
  const { mobile } = useDisplay()

  const menu = ref(false)
  const connectedBranchIds = ref([])
  let unsubscribe = null

  // Without a live socket we have no idea which branches are actually online,
  // so the count would be stale -- surface the outage message instead of a number.
  const serverDown = computed(() => status.value !== 'open')
  const statusColor = computed(() => {
    if (serverDown.value) return 'white'
    return connectedBranchIds.value.length > 0 ? 'success' : 'error'
  })

  function branchName (branchId) {
    return branches.value.find(b => b.id === branchId)?.name ?? `Branch #${branchId}`
  }

  onMounted(async () => {
    await fetchBranches()
    try {
      connectedBranchIds.value = await fetchConnectedBranches()
    } catch {
      // Informational only -- the WS push still keeps this current from
      // here on even if the initial fetch failed.
    }
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (message.type === 'branch_connection_changed') {
        connectedBranchIds.value = message.branch_ids ?? []
      }
    })
  })

  onUnmounted(() => {
    unsubscribe?.()
  })
</script>

<style scoped>
  /* A plain white fill vanishes on light backgrounds, so ring the dot when the
     server is unreachable to keep it legible in both themes. */
  .status-dot--outlined {
    border: 1px solid rgba(var(--v-border-color), 0.4);
    border-radius: 50%;
  }

  /* On mobile the pill collapses to just the status dot -- fix it to the same
     square SyncCard's icon-only pill uses so the two sit level. */
  .icon-pill {
    width: 34px;
    height: 34px;
    justify-content: center;
  }
</style>
