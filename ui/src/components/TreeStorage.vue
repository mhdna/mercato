<template>
  <div class="d-flex" style="height: 100%">
    <v-sheet class="flex-shrink-0" width="240">
      <v-treeview
        v-model:activated="activated"
        activatable
        :items="items"
        item-value="id"
        open-on-click
      >
        <template #prepend="{ item }">
          <v-icon :icon="item.icon" />
        </template>
      </v-treeview>
    </v-sheet>

    <v-divider vertical />

    <div class="flex-grow-1 pa-4" style="overflow-y: auto">
      <template v-if="activeNode === 'expenses'">
        <div class="d-flex align-center justify-space-between mb-4">
          <div class="text-h6">Expense receipts</div>
          <v-progress-circular v-if="loading" indeterminate size="20" />
        </div>

        <v-alert v-if="loadError" class="mb-4" type="error" variant="tonal">{{ loadError }}</v-alert>

        <div v-if="!loading && images.length === 0" class="text-medium-emphasis">
          No receipt photos uploaded yet.
        </div>

        <div class="image-grid">
          <v-card
            v-for="image in images"
            :key="image.id"
            class="image-card"
            hover
            @click="openPreview(image)"
          >
            <v-img aspect-ratio="1" cover :src="thumbnailUrls[image.id]">
              <template #placeholder>
                <div class="d-flex align-center justify-center fill-height">
                  <v-progress-circular indeterminate size="24" />
                </div>
              </template>
            </v-img>
            <v-card-text class="pa-2">
              <div class="text-body-2 text-truncate">{{ image.expense_description }}</div>
              <div class="text-caption text-medium-emphasis">{{ image.branch_name }}</div>
              <div class="text-caption text-medium-emphasis">{{ formatDate(image.created_at) }}</div>
            </v-card-text>
          </v-card>
        </div>

        <div v-if="!loading && images.length < total" class="text-center mt-4">
          <v-btn variant="text" @click="loadMore">Load more</v-btn>
        </div>
      </template>

      <template v-else>
        <div class="text-medium-emphasis">Select a folder on the left.</div>
      </template>
    </div>

    <v-dialog v-model="previewDialog" max-width="720">
      <v-card>
        <v-img v-if="previewImage" contain max-height="80vh" :src="thumbnailUrls[previewImage.id]" />
        <v-card-text v-if="previewImage">
          <div class="text-subtitle-1">{{ previewImage.expense_description }}</div>
          <div class="text-caption text-medium-emphasis">
            {{ previewImage.branch_name }} — {{ formatDate(previewImage.created_at) }}
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="previewDialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
  import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
  import { useAdminSocket } from '@/composables/useAdminSocket'
  import { useExpenseImages } from '@/composables/useExpenseImages'

  // Only one node today (per CLAUDE.md-style room-to-grow convention used
  // elsewhere in this codebase) -- more storage kinds land as siblings
  // here later without reshaping this component.
  const items = [
    { id: 'expenses', title: 'Expenses', icon: 'mdi-receipt-text-check' },
  ]

  const activated = ref(['expenses'])
  const activeNode = computed(() => activated.value[0] ?? null)

  const { fetchExpenseImages, fetchExpenseImageObjectUrl } = useExpenseImages()
  const { ensureConnected, onMessage } = useAdminSocket()

  const images = ref([])
  const total = ref(0)
  const pageId = ref(0)
  const pageSize = 24
  const loading = ref(false)
  const loadError = ref('')
  const thumbnailUrls = ref({})

  const previewDialog = ref(false)
  const previewImage = ref(null)
  let unsubscribe = null

  function formatDate (value) {
    return new Date(value).toLocaleString()
  }

  async function loadThumbnails (rows) {
    await Promise.all(rows.map(async image => {
      if (thumbnailUrls.value[image.id]) return
      try {
        thumbnailUrls.value[image.id] = await fetchExpenseImageObjectUrl(image.id)
      } catch {
        // Leave it unset -- the v-img placeholder just stays put.
      }
    }))
  }

  async function loadImages ({ reset = false } = {}) {
    loading.value = true
    loadError.value = ''
    try {
      const page = reset ? 0 : pageId.value
      const data = await fetchExpenseImages(page, pageSize)
      images.value = reset ? data.images : [...images.value, ...data.images]
      total.value = data.total
      pageId.value = page + 1
      await loadThumbnails(data.images)
    } catch (err) {
      loadError.value = err.message
    } finally {
      loading.value = false
    }
  }

  function loadMore () {
    loadImages()
  }

  function openPreview (image) {
    previewImage.value = image
    previewDialog.value = true
  }

  watch(activeNode, node => {
    if (node === 'expenses' && images.value.length === 0) {
      loadImages({ reset: true })
    }
  })

  onMounted(() => {
    if (activeNode.value === 'expenses') {
      loadImages({ reset: true })
    }
    ensureConnected()
    unsubscribe = onMessage(message => {
      if (message.type === 'branch_expense_images_uploaded') {
        pageId.value = 0
        loadImages({ reset: true })
      }
    })
  })

  onUnmounted(() => {
    unsubscribe?.()
    Object.values(thumbnailUrls.value).forEach(url => URL.revokeObjectURL(url))
  })
</script>

<style scoped>
.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.image-card {
  cursor: pointer;
}
</style>
