<template>
  <main class="home-empty-state">
    <div class="home-content">
      <div class="home-top">
        <section class="home-hero text-center">
          <div
            v-if="mdAndUp"
            aria-label="GB"
            class="home-desktop-logo mx-auto"
            role="img"
            :style="{ '--home-logo-color': logoColor, '--home-logo-image': `url(${gbLogo})` }"
          />
          <GbLogo v-else class="mx-auto" />
          <h1 class="mt-5 text-h5 font-weight-medium">Welcome to GB Cloud</h1>
          <p class="mt-2 text-body-2 text-medium-emphasis">
            Choose where you would like to start.
          </p>
        </section>

        <section aria-label="Quick access" class="home-shortcuts">
          <v-card
            v-for="item in shortcuts"
            :key="item.to"
            class="home-shortcut"
            rounded="lg"
            :to="item.to"
            variant="outlined"
          >
            <v-card-text class="d-flex align-center pa-4">
              <v-avatar class="me-3" color="surface-light" rounded="lg" size="42">
                <v-icon class="text-medium-emphasis" :icon="item.icon" size="22" />
              </v-avatar>
              <div class="min-width-0">
                <div class="text-body-1 font-weight-medium">{{ item.title }}</div>
                <div class="shortcut-description text-caption text-medium-emphasis">
                  {{ item.description }}
                </div>
              </div>
              <v-icon class="ms-auto text-medium-emphasis" icon="mdi-chevron-right" size="20" />
            </v-card-text>
          </v-card>
        </section>

        <v-autocomplete
          v-model="selectedLink"
          auto-select-first
          autocomplete="off"
          class="home-search"
          clearable
          density="comfortable"
          hide-details
          item-title="title"
          item-value="to"
          :items="links"
          :menu-props="searchMenuProps"
          no-data-text="No links found"
          placeholder="Search links..."
          prepend-inner-icon="mdi-magnify"
          rounded="pill"
          variant="outlined"
          @update:model-value="openLink"
        />
      </div>

      <section aria-label="Recent activity" class="home-activities">
        <div class="home-activities-head d-flex align-center px-1 mb-2">
          <v-icon class="me-2 text-medium-emphasis" icon="mdi-history" size="18" />
          <span class="text-body-2 font-weight-medium">Recent activity</span>
          <v-progress-circular
            v-if="loading && activities.length === 0"
            class="ms-2"
            indeterminate
            size="14"
            width="2"
          />
        </div>

        <v-card class="home-activities-card" rounded="lg" variant="outlined">
          <v-alert
            v-if="error"
            class="ma-2"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ error }}
          </v-alert>

          <div v-else class="home-activities-scroll">
            <v-list class="py-1" density="compact" lines="two">
              <v-list-item
                v-for="(item, index) in activities"
                :key="index"
                class="px-3 activity-row"
                @click="openDetail(item)"
              >
                <template #prepend>
                  <v-avatar
                    class="me-3"
                    :color="visual(item).color"
                    rounded="lg"
                    size="34"
                    variant="tonal"
                  >
                    <v-icon :icon="visual(item).icon" size="18" />
                  </v-avatar>
                </template>
                <v-list-item-title class="text-body-2 font-weight-medium">
                  {{ item.activity || 'Activity' }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption">
                  {{ [item.branch_name, item.details].filter(Boolean).join(' · ') || '—' }}
                </v-list-item-subtitle>
                <template #append>
                  <div class="text-right">
                    <div
                      v-if="hasAmount(item)"
                      class="text-body-2 font-weight-medium"
                      :class="amountClass(item.amount)"
                    >
                      {{ money(item.amount, item.currency_code) }}
                    </div>
                    <div class="text-caption text-medium-emphasis">
                      {{ dateTime(item.synced_at) }}
                    </div>
                  </div>
                </template>
              </v-list-item>
            </v-list>

            <div
              v-if="!loading && activities.length === 0"
              class="text-body-2 text-medium-emphasis py-4 text-center"
            >
              No recent activity yet.
            </div>
          </div>
        </v-card>
      </section>
    </div>

    <v-dialog v-model="detailOpen" max-width="460">
      <v-card v-if="detail" rounded="lg">
        <v-card-title class="d-flex align-center ga-3 px-4 py-3">
          <v-avatar :color="visual(detail).color" rounded="lg" size="38" variant="tonal">
            <v-icon :icon="visual(detail).icon" size="20" />
          </v-avatar>
          <span class="text-subtitle-1 font-weight-medium">{{ detail.activity || 'Activity' }}</span>
          <v-spacer />
          <v-btn
            aria-label="Close"
            icon="mdi-close"
            size="small"
            variant="text"
            @click="detailOpen = false"
          />
        </v-card-title>
        <v-divider />
        <v-list class="py-2" density="compact">
          <v-list-item>
            <v-list-item-subtitle class="text-caption">Branch</v-list-item-subtitle>
            <v-list-item-title class="text-body-2">{{ detail.branch_name || '—' }}</v-list-item-title>
          </v-list-item>
          <v-list-item>
            <v-list-item-subtitle class="text-caption">Details</v-list-item-subtitle>
            <v-list-item-title class="text-body-2" style="white-space: normal">
              {{ detail.details || '—' }}
            </v-list-item-title>
          </v-list-item>
          <v-list-item v-if="hasAmount(detail)">
            <v-list-item-subtitle class="text-caption">Amount</v-list-item-subtitle>
            <v-list-item-title class="text-body-2" :class="amountClass(detail.amount)">
              {{ money(detail.amount, detail.currency_code) }}
            </v-list-item-title>
          </v-list-item>
          <v-list-item>
            <v-list-item-subtitle class="text-caption">Synced at</v-list-item-subtitle>
            <v-list-item-title class="text-body-2">{{ dateTime(detail.synced_at) }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-card>
    </v-dialog>
  </main>
</template>

<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { useDisplay, useTheme } from 'vuetify'
  import gbLogo from '@/assets/gb-logo.png'
  import GbLogo from '@/components/GbLogo.vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { authFetch } from '@/composables/useApi'
  import { BRANCH_ACTIVITY_TYPES } from '@/composables/useBranchActivityMessage'
  import { API_BASE } from '@/config'
  import { appendItems, navItems } from '@/data/navItems'

  const router = useRouter()
  const { mdAndUp } = useDisplay()
  const theme = useTheme()
  const logoColor = computed(() => theme.global.current.value.dark
    ? 'rgb(var(--v-theme-on-surface))'
    : 'rgba(var(--v-theme-on-surface), 0.7)')
  const selectedLink = ref(null)
  const searchMenuProps = {
    location: 'bottom',
    maxHeight: 152,
    offset: 4,
  }

  const links = []
  for (const item of [...navItems, ...appendItems]) {
    if (item.children?.length > 0) {
      links.push(...item.children)
    } else if (item.to) {
      links.push(item)
    }
  }

  function openLink (to) {
    if (!to) return
    router.push(to)
    selectedLink.value = null
  }

  const shortcuts = [
    { title: 'Dashboard', description: 'View sales, expenses, and totals', icon: 'mdi-view-dashboard', to: '/dashboard' },
    { title: 'Invoices', description: 'Review retail and wholesale invoices', icon: 'mdi-invoice', to: '/invoices' },
    { title: 'Inventory', description: 'Check products and stock levels', icon: 'mdi-package-variant', to: '/inventory' },
    { title: 'Expenses', description: 'Add and review expenses', icon: 'mdi-currency-usd-off', to: '/expenses' },
    { title: 'Branches', description: 'Manage branches and check sync status', icon: 'mdi-store-outline', to: '/branches' },
    { title: 'Settings', description: 'Change app settings', icon: 'mdi-cog-outline', to: '/settings' },
  ]

  // Recent activity: a read-only glance at the latest branch-sync activity.
  // Reuses the dashboard's /activities endpoint, which defaults to the
  // current month ordered newest-first when no filters are passed. This is a
  // fixed, short feed -- just the newest page -- not a paginated log: the
  // full history lives on the dashboard. The panel fills the bottom third of
  // the (non-scrolling) home page and scrolls internally, and a click on any
  // row opens a details dialog. It refreshes itself off the shared admin
  // websocket (the same push stream the dashboard listens to), so there's no
  // manual refresh button.
  const PAGE_SIZE = 20

  const activities = ref([])
  const loading = ref(false)
  const error = ref('')

  const detailOpen = ref(false)
  const detail = ref(null)

  function openDetail (item) {
    detail.value = item
    detailOpen.value = true
  }

  // Per-activity-type colour + glyph. `activity` is the union label the SQL
  // feed stamps on every row (see DashboardActivitiesList); invoices and
  // loans additionally swing colour on the sign of the amount.
  const ACTIVITY_VISUALS = {
    'Invoice': { icon: 'mdi-sale', color: 'success' },
    'Expense': { icon: 'mdi-cash-minus', color: 'error' },
    'Loan': { icon: 'mdi-hand-coin', color: 'warning' },
    'Shift closed': { icon: 'mdi-cash-register', color: 'info' },
    'Settlement changed': { icon: 'mdi-cash-edit', color: 'purple' },
    'Attendance': { icon: 'mdi-fingerprint', color: 'primary' },
    'Attendance changed': { icon: 'mdi-clock-edit-outline', color: 'primary' },
  }

  function visual (item) {
    const base = ACTIVITY_VISUALS[item.activity] ?? { icon: 'mdi-pulse', color: 'info' }
    if ((item.activity === 'Invoice' || item.activity === 'Loan') && item.amount < 0) {
      return { ...base, color: 'error' }
    }
    return base
  }

  // Fetch the newest page and replace the list. There's no "load more" --
  // the dashboard is where the full, filterable history lives.
  async function reload () {
    loading.value = true
    error.value = ''
    try {
      const response = await authFetch(
        `${API_BASE}/dashboard/activities?page_size=${PAGE_SIZE}&page_id=0`,
      )
      if (!response.ok) {
        const body = await response.json().catch(() => ({}))
        throw new Error(body.error || `Failed to load recent activity (${response.status})`)
      }
      const data = await response.json()
      activities.value = data.activities ?? []
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  }

  // A branch catch-up can fire a burst of pushes; debounce so the feed
  // reloads once after the dust settles.
  const { ensureConnected, onMessage } = useAdminSocket()
  let refreshTimer = null
  ensureConnected()
  const unsubscribe = onMessage(message => {
    if (!BRANCH_ACTIVITY_TYPES.includes(message?.type)) return
    clearTimeout(refreshTimer)
    refreshTimer = setTimeout(reload, 500)
  })

  function money (amount, currency = '') {
    const value = Number(amount || 0) / 100
    const formatted = new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(Math.abs(value))
    return `${amount < 0 ? '−' : ''}${formatted} ${currency}`.trim()
  }

  function hasAmount (item) {
    return item.amount != null && item.amount !== 0 && !String(item.activity).startsWith('Attendance')
  }

  function amountClass (amount) {
    if (amount > 0) return 'text-success'
    if (amount < 0) return 'text-error'
    return 'text-medium-emphasis'
  }

  function dateTime (value) {
    return value ? new Date(value).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' }) : '—'
  }

  onMounted(reload)
  onUnmounted(() => {
    clearTimeout(refreshTimer)
    unsubscribe()
  })
</script>

<style scoped>
  .home-empty-state {
    display: flex;
    justify-content: center;
    width: 100%;
    height: 100%;
    min-height: 0;
    padding: clamp(16px, 4vh, 40px) 24px 24px;
    overflow: hidden;
  }

  /* The page is a fixed-height, non-scrolling column: the hero/shortcuts/
     search block takes the top two thirds, the recent-activity panel the
     bottom third and scrolls within itself. */
  .home-content {
    display: flex;
    flex-direction: column;
    width: 100%;
    min-height: 0;
  }

  .home-top {
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
  }

  .home-hero {
    margin-inline: auto;
    max-width: 520px;
  }

  .home-desktop-logo {
    width: clamp(72px, 9vw, 132px);
    aspect-ratio: 1;
    background-color: var(--home-logo-color);
    mask: var(--home-logo-image) center / contain no-repeat;
    -webkit-mask: var(--home-logo-image) center / contain no-repeat;
  }

  .home-search {
    margin: 20px auto 0;
    max-width: 620px;
  }

  .home-activities {
    display: flex;
    flex: 0 0 33.333%;
    flex-direction: column;
    min-height: 0;
    max-width: 620px;
    margin: 16px auto 0;
    width: 100%;
  }

  .home-activities-card {
    display: flex;
    flex: 1 1 0;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .home-activities-scroll {
    flex: 1 1 0;
    min-height: 0;
    overflow-y: auto;
  }

  .activity-row {
    cursor: pointer;
  }

  .home-shortcuts {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
    max-width: 980px;
    margin: 24px auto 0;
  }

  .home-shortcut {
    display: flex;
    min-height: 84px;
    border-color: rgb(var(--v-theme-surface-light));
    transition: border-color 0.15s ease;
  }

  .home-shortcut:hover {
    border-color: rgba(var(--v-theme-primary), 0.55);
  }

  .min-width-0 {
    min-width: 0;
  }

  .shortcut-description {
    line-height: 1.4;
    white-space: normal;
  }

  @media (max-width: 900px) {
    .home-shortcuts {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 600px) {
    .home-empty-state {
      padding-inline: 8px;
    }

    .home-shortcuts {
      grid-template-columns: 1fr;
      margin-top: 32px;
    }
  }
</style>
