<template>
  <main class="home-empty-state">
    <div class="home-content">
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

      <section aria-label="Recent activity" class="home-activities">
        <div class="home-activities-head d-flex align-center px-1 mb-2">
          <v-icon class="me-2 text-medium-emphasis" icon="mdi-history" size="18" />
          <span class="text-body-2 font-weight-medium">Recent activity</span>
          <v-progress-circular
            v-if="activitiesLoading"
            class="ms-2"
            indeterminate
            size="14"
            width="2"
          />
        </div>

        <v-card rounded="lg" variant="outlined">
          <v-alert
            v-if="activitiesError"
            class="ma-2"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ activitiesError }}
          </v-alert>

          <v-list v-else class="py-1" density="compact" lines="two">
            <template v-if="activities.length > 0">
              <v-list-item
                v-for="(item, index) in activities"
                :key="index"
                class="px-3"
              >
                <template #prepend>
                  <v-avatar class="me-3" color="surface-light" rounded="lg" size="34">
                    <v-icon class="text-medium-emphasis" icon="mdi-pulse" size="18" />
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
                    <div v-if="item.amount" class="text-body-2 font-weight-medium">
                      {{ money(item.amount, item.currency_code) }}
                    </div>
                    <div class="text-caption text-medium-emphasis">
                      {{ dateTime(item.synced_at) }}
                    </div>
                  </div>
                </template>
              </v-list-item>
            </template>
            <v-list-item v-else-if="!activitiesLoading" class="px-3">
              <v-list-item-title class="text-body-2 text-medium-emphasis">
                No recent activity yet.
              </v-list-item-title>
            </v-list-item>
            <v-list-item v-else class="px-3">
              <v-list-item-title class="text-body-2 text-medium-emphasis">
                Loading…
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-card>
      </section>
    </div>
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

  // Recent activity: a small, read-only feed of what the branches have
  // synced lately. Reuses the dashboard's /activities endpoint, which
  // defaults to the current month ordered newest-first when no filters
  // are passed. It refreshes itself off the shared admin websocket -- the
  // same push stream the dashboard's Recent Activities tab listens to --
  // so there's no manual refresh button.
  const activities = ref([])
  const activitiesLoading = ref(false)
  const activitiesError = ref('')

  let requestId = 0
  async function loadActivities () {
    const current = ++requestId
    activitiesLoading.value = true
    activitiesError.value = ''
    try {
      const response = await authFetch(`${API_BASE}/dashboard/activities?page_size=8&page_id=0`)
      if (!response.ok) {
        const body = await response.json().catch(() => ({}))
        throw new Error(body.error || `Failed to load recent activity (${response.status})`)
      }
      const data = await response.json()
      if (current === requestId) activities.value = data.activities ?? []
    } catch (error) {
      if (current === requestId) activitiesError.value = error.message
    } finally {
      if (current === requestId) activitiesLoading.value = false
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
    refreshTimer = setTimeout(loadActivities, 500)
  })

  function money (amount, currency = '') {
    const value = Number(amount || 0) / 100
    const formatted = new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(Math.abs(value))
    return `${amount < 0 ? '−' : ''}${formatted} ${currency}`.trim()
  }

  function dateTime (value) {
    return value ? new Date(value).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' }) : '—'
  }

  onMounted(loadActivities)
  onUnmounted(() => {
    clearTimeout(refreshTimer)
    unsubscribe()
  })
</script>

<style scoped>
  .home-empty-state {
    display: flex;
    align-items: flex-start;
    justify-content: center;
    width: 100%;
    height: 100%;
    min-height: 0;
    /* Sit the content in the upper third rather than dead-centre so the
       recent-activity feed below the search bar stays in view. */
    padding: clamp(32px, 9vh, 96px) 24px 32px;
    overflow-y: auto;
  }

  .home-content {
    width: 100%;
  }

  .home-hero {
    margin-inline: auto;
    max-width: 520px;
  }

  .home-desktop-logo {
    width: clamp(100px, 13vw, 180px);
    aspect-ratio: 1;
    background-color: var(--home-logo-color);
    mask: var(--home-logo-image) center / contain no-repeat;
    -webkit-mask: var(--home-logo-image) center / contain no-repeat;
  }

  .home-search {
    margin: 24px auto 0;
    max-width: 620px;
  }

  .home-activities {
    margin: 20px auto 0;
    max-width: 620px;
  }

  .home-shortcuts {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
    max-width: 980px;
    margin: 48px auto 0;
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
