<template>
  <v-snackbar-queue
    v-model="toasts"
    color="surface-light"
    location="bottom end"
    max-width="380"
    vertical
  >
    <template #text="{ item }">
      <div class="d-flex align-center ga-2 text-body-large">
        <v-icon :icon="item.icon" size="20" />
        <div v-if="item.money">
          {{ item.branchName }}:
          <v-icon
            :class="textColorClass(item)"
            :icon="item.trendIcon"
            size="18"
            style="vertical-align: text-bottom"
          />
          <span :class="textColorClass(item)">{{ item.amountText }}</span> {{ item.currencyCode }} {{ item.kindWord }}
        </div>
        <div v-else>{{ item.text }}</div>
      </div>
    </template>
    <template #actions="{ props }">
      <v-btn color="indigo" variant="text" v-bind="props">
        Close
      </v-btn>
    </template>
  </v-snackbar-queue>
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

  const toasts = ref([])
  let unsubscribe = null

  function branchName (branchId) {
    return branches.value.find(b => b.id === branchId)?.name ?? `Branch #${branchId}`
  }

  function textColorClass (item) {
    if (item.amount > 0) return 'text-green'
    if (item.amount < 0) return 'text-red'
    return 'text-yellow'
  }

  function pushToast (message) {
    const described = describeBranchActivity(message, branchName)
    if (!described) return
    toasts.value.push({
      icon: described.icon,
      money: described.money,
      text: described.text,
      // Money-layout fields (only read when described.money is true).
      trendIcon: described.trendIcon,
      branchName: branchName(message.branch_id),
      amountText: formatMoney(message.amount),
      currencyCode: message.currency_code,
      kindWord: described.kindWord,
      amount: message.amount,
      timeout: settingsStore.activityMessageSeconds * 1000,
    })
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
    unsubscribe?.()
  })
</script>
