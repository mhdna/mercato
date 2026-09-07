<template>
  <!-- A fixed reel pinned to the bottom-right corner. Up to MAX_VISIBLE
       toasts sit side by side, newest nearest the corner; extras queue and
       slot in as older ones dismiss. Replaces v-snackbar-queue, which only
       ever showed one at a time. -->
  <div class="activity-reel">
    <TransitionGroup class="d-flex align-end ga-2" name="toast" tag="div">
      <v-card
        v-for="item in activeItems"
        :key="item.id"
        class="activity-toast d-flex align-center ga-2 pa-3"
        color="surface-light"
      >
        <v-icon :icon="item.icon" size="20" />
        <div class="text-body-2">
          <template v-if="item.money">
            {{ item.branchName }}:
            <v-icon
              :class="item.colorClass"
              :icon="item.trendIcon"
              size="18"
              style="vertical-align: text-bottom"
            />
            <span :class="item.colorClass">{{ item.amountText }}</span> {{ item.currencyCode }} {{ item.kindWord }}
          </template>
          <template v-else>{{ item.text }}</template>
        </div>
        <v-btn
          class="ms-1"
          density="comfortable"
          icon="mdi-close"
          size="x-small"
          variant="text"
          @click="removeItem(item.id)"
        />
      </v-card>
    </TransitionGroup>
  </div>
</template>

<script setup>
  import { onMounted, onUnmounted, ref } from 'vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { describeBranchActivity, isBranchActivityMessage } from '@/composables/useBranchActivityMessage'
  import { useBranches } from '@/composables/useBranches'
  import { useSettingsStore } from '@/stores/settings'
  import { formatMoney } from '@/utils/money'

  const { branches, fetchBranches } = useBranches()
  const { ensureConnected, onMessage } = useAdminSocket()
  const settingsStore = useSettingsStore()

  const MAX_VISIBLE = 3
  const queue = ref([])
  const activeItems = ref([]) // oldest first; newest renders nearest the corner
  const itemTimers = new Map()
  let itemSeq = 0
  let unsubscribe = null

  function branchName (branchId) {
    return branches.value.find(b => b.id === branchId)?.name ?? `Branch #${branchId}`
  }

  function amountColorClass (amount) {
    if (amount > 0) return 'text-green'
    if (amount < 0) return 'text-red'
    return 'text-yellow'
  }

  function scheduleRemoval (id) {
    itemTimers.set(id, setTimeout(() => removeItem(id), settingsStore.activityMessageSeconds * 1000))
  }

  function removeItem (id) {
    clearTimeout(itemTimers.get(id))
    itemTimers.delete(id)
    const i = activeItems.value.findIndex(it => it.id === id)
    if (i !== -1) activeItems.value.splice(i, 1)
    if (queue.value.length > 0 && activeItems.value.length < MAX_VISIBLE) {
      const next = queue.value.shift()
      activeItems.value.push(next)
      scheduleRemoval(next.id)
    }
  }

  function pushToast (message) {
    const described = describeBranchActivity(message, branchName)
    if (!described) return
    const item = {
      id: ++itemSeq,
      icon: described.icon,
      money: described.money,
      text: described.text,
      // Money-layout fields (only read when described.money is true).
      trendIcon: described.trendIcon,
      branchName: branchName(message.branch_id),
      amountText: formatMoney(message.amount),
      currencyCode: message.currency_code,
      kindWord: described.kindWord,
      colorClass: amountColorClass(message.amount),
    }
    if (activeItems.value.length < MAX_VISIBLE) {
      activeItems.value.push(item)
      scheduleRemoval(item.id)
    } else {
      queue.value.push(item)
    }
  }

  onMounted(() => {
    fetchBranches()
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (settingsStore.activityDisplayMode === 'notification' && isBranchActivityMessage(message)) {
        pushToast(message)
      }
    })
  })

  onUnmounted(() => {
    for (const timer of itemTimers.values()) clearTimeout(timer)
    unsubscribe?.()
  })
</script>

<style scoped>
.activity-reel {
  position: fixed;
  right: 16px;
  bottom: 16px;
  z-index: 3000; /* same layer v-snackbar used */
  pointer-events: none;
}

.activity-toast {
  max-width: 340px;
  pointer-events: auto;
}

.toast-enter-active,
.toast-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  transform: translateY(12px);
  opacity: 0;
}

/* Take a dismissed toast out of flow so its neighbours slide over. */
.toast-leave-active {
  position: absolute;
}

.toast-move {
  transition: transform 0.3s ease;
}
</style>
