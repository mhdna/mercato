<template>
  <v-menu v-model="menu" :close-on-content-click="false" location="bottom center" :offset="8">
    <template #activator="{ props }">
      <v-card
        v-bind="props"
        class="d-flex align-center py-1 px-2"
        rounded="xl"
        style="cursor: pointer"
        variant="tonal"
      >
        <v-icon :color="connectedBranchIds.length > 0 ? 'success' : 'error'" icon="mdi-circle" size="16" />
        <span class="ms-1 text-body-2">
          {{ connectedBranchIds.length }} {{ connectedBranchIds.length === 1 ? 'device' : 'devices' }} online
        </span>
      </v-card>
    </template>

    <v-card max-width="320" min-width="240">
      <v-card-title class="text-subtitle-1">Connected branches</v-card-title>
      <v-divider />

      <v-list v-if="connectedBranchIds.length > 0" density="compact">
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
  import { onMounted, onUnmounted, ref } from 'vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { useBranches } from '@/composables/useBranches'

  const { branches, fetchBranches, fetchConnectedBranches } = useBranches()
  const { ensureConnected, onMessage } = useAdminSocket()

  const menu = ref(false)
  const connectedBranchIds = ref([])
  let unsubscribe = null

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
