<template>
  <main class="home-empty-state">
    <section class="home-hero text-center">
      <GbLogo class="mx-auto" />
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
  </main>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import { useRouter } from 'vue-router'
  import GbLogo from '@/components/GbLogo.vue'
  import { appendItems, navItems } from '@/data/navItems'

  const router = useRouter()
  const selectedLink = ref(null)
  const searchMenuProps = {
    location: 'bottom',
    maxHeight: 152,
    offset: 4,
  }

  const links = []
  for (const item of [...navItems, ...appendItems]) {
    if (item.children?.length) {
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
</script>

<style scoped>
  .home-empty-state {
    width: 100%;
    height: 100%;
    min-height: 0;
    padding: clamp(32px, 7vh, 72px) 24px 40px;
    overflow-y: auto;
  }

  .home-hero {
    margin-inline: auto;
    max-width: 520px;
  }

  .home-search {
    margin: 24px auto 0;
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
