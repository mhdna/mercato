<template>
  <!-- Single root so the flex sizing set on <DashboardList> actually lands
    somewhere -- a multi-root component silently drops fallthrough style. -->
  <div class="dash-list">
    <v-card class="dash-card" variant="outlined">
      <div class="dash-card-header d-flex align-center px-2 py-1">
        <v-icon class="me-2 text-medium-emphasis" :icon="icon" size="18" />
        <span class="text-caption font-weight-medium text-uppercase text-medium-emphasis">
          {{ title }}
        </span>
        <v-spacer />
        <span v-if="total >= 0" class="text-caption text-disabled">{{ total }}</span>
      </div>

      <v-alert
        v-if="error"
        class="ma-1"
        density="compact"
        type="error"
        variant="tonal"
      >
        {{ error }}
        <template #append>
          <v-btn size="x-small" variant="text" @click="reset">Retry</v-btn>
        </template>
      </v-alert>

      <v-virtual-scroll
        v-else
        ref="scroller"
        class="dash-scroll"
        :height="height || undefined"
        :item-height="rowHeight"
        :items="items"
        @scroll.passive="onScroll"
      >
        <template #default="{ item }">
          <v-menu
            :close-delay="60"
            location="start"
            offset="4"
            :open-delay="140"
            open-on-hover
          >
            <template #activator="{ props: hover }">
              <div
                class="dash-row d-flex align-center px-2"
                :style="{ height: `${rowHeight}px` }"
                v-bind="hover"
              >
                <slot :item="item" name="row" />
              </div>
            </template>

            <v-card class="pa-3" max-width="300" min-width="220">
              <slot :item="item" name="details" />
              <v-btn
                append-icon="mdi-open-in-new"
                block
                class="mt-1"
                size="small"
                variant="tonal"
                @click="openItem = item"
              >
                Open
              </v-btn>
            </v-card>
          </v-menu>
        </template>
      </v-virtual-scroll>

      <div
        v-if="!error && loading && items.length === 0"
        class="flex-grow-1 d-flex align-center justify-center"
      >
        <v-progress-circular color="primary" indeterminate size="22" width="2" />
      </div>
      <div
        v-else-if="!error && !loading && items.length === 0"
        class="flex-grow-1 d-flex align-center justify-center text-medium-emphasis text-body-2"
      >
        Nothing here yet
      </div>
    </v-card>

    <v-dialog v-model="dialog" max-width="460">
      <v-card v-if="openItem">
        <v-card-title class="d-flex align-center text-body-1">
          <v-icon class="me-2" :icon="icon" size="20" />
          {{ title }} detail
        </v-card-title>
        <v-card-text>
          <slot :item="openItem" name="details" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Close" @click="dialog = false" />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref, watch } from 'vue'
  import { dedupedFetch } from '@/composables/useRequestDedup'

  const props = defineProps({
    // Base URL with filters already applied; DashboardList adds paging.
    apiUrl: {
      type: String,
      required: true,
    },
    rootKey: {
      type: String,
      required: true,
    },
    title: {
      type: String,
      default: '',
    },
    icon: {
      type: String,
      default: '',
    },
    // Fixed body height. Omit (default) to flex-fill the available space.
    height: {
      type: [Number, String],
      default: null,
    },
    pageSize: {
      type: Number,
      default: 25,
    },
    // Hard ceiling on rows kept in memory. A dashboard glance never needs
    // to scroll thousands, and v-virtual-scroll already keeps the DOM tiny,
    // so this just bounds the JS array + stops paging the DB forever.
    maxItems: {
      type: Number,
      default: 300,
    },
    // Optional local rows for dashboard previews and empty API environments.
    mockItems: {
      type: Array,
      default: null,
    },
  })

  const rowHeight = 46
  const items = ref([])
  const total = ref(-1)
  const loading = ref(false)
  const error = ref('')
  const scroller = ref(null)

  let offset = 0
  let done = false
  let seq = 0

  async function fetchPage (reset) {
    if (loading.value) return
    if (!reset && done) return

    if (props.mockItems !== null) {
      items.value = props.mockItems
      total.value = props.mockItems.length
      offset = props.mockItems.length
      done = true
      return
    }

    loading.value = true
    error.value = ''
    const mySeq = ++seq
    const from = reset ? 0 : offset
    try {
      const url = new URL(props.apiUrl)
      url.searchParams.set('page_size', props.pageSize)
      url.searchParams.set('page_id', from)
      const data = await dedupedFetch(url.toString())
      if (mySeq !== seq) return
      const batch = Array.isArray(data) ? data : (data[props.rootKey] ?? [])
      total.value = Array.isArray(data) ? -1 : (data.total ?? -1)
      items.value = reset ? batch : [...items.value, ...batch]
      offset = items.value.length
      done = batch.length < props.pageSize || items.value.length >= props.maxItems
    } catch (error_) {
      if (mySeq === seq) error.value = error_.message
    } finally {
      if (mySeq === seq) loading.value = false
    }
  }

  function onScroll (event) {
    const el = event.target
    if (el && el.scrollTop + el.clientHeight >= el.scrollHeight - 200) {
      fetchPage(false)
    }
  }

  function reset () {
    done = false
    offset = 0
    const el = scroller.value?.$el
    if (el) el.scrollTop = 0
    fetchPage(true)
  }

  const openItem = ref(null)
  const dialog = computed({
    get: () => openItem.value != null,
    set: value => {
      if (!value) openItem.value = null
    },
  })

  watch(() => props.apiUrl, reset)
  onMounted(() => fetchPage(true))

  defineExpose({ reset })
</script>

<style scoped>
  /* Fills the flex track it's placed in -- the sizing (flex + height)
     comes through as a fallthrough style on this root -- and lays its card
     out as a column so the scroller can take the leftover height. */
  .dash-list {
    display: flex;
    flex-direction: column;
    min-width: 0;
    height: 100%;
  }

  .dash-card {
    display: flex;
    flex: 1 1 0;
    flex-direction: column;
    min-height: 0;
  }

  .dash-card.v-card--variant-outlined {
    border-color: rgb(var(--v-theme-surface-light));
  }

  .dash-card-header {
    background-color: rgb(var(--v-theme-surface-light));
  }

  /* Grow to fill the card, but allow shrinking below content so it scrolls
     instead of pushing the card taller. */
  .dash-scroll {
    flex: 1 1 0;
    min-height: 0;
    overflow-y: auto;
  }

  .dash-row {
    border-radius: 6px;
    transition: background-color 0.12s ease;
  }

  .dash-row:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.04);
  }
</style>
