<script setup>
  import { computed, ref } from 'vue'
  import AssetTreemap from '@/components/Stats/AssetTreemap.vue'
  import { assetLocationOptions, assetTotals } from '@/data/assetValues'

  const options = ref([...assetLocationOptions])
  const selected = ref('all')

  const totals = computed(() => assetTotals(selected.value))
  const totalValue = computed(() => '$ ' + totals.value.value.toLocaleString('en-US'))
  const totalCount = computed(() => totals.value.qty.toLocaleString('en-US') + ' units')

  const dialog = ref(false)
  const newName = ref('')

  function openAdd () {
    newName.value = ''
    dialog.value = true
  }

  function addLocation () {
    const name = newName.value.trim()
    if (!name) return
    const value = name.toLowerCase().replace(/\s+/g, '-')
    options.value.push({ title: name, value })
    selected.value = value
    dialog.value = false
  }
</script>

<template>
  <div class="assets-treemap-page">
    <div class="filters d-flex align-center ga-3 pa-3">
      <v-select
        v-model="selected"
        density="compact"
        hide-details
        :items="options"
        label="Location"
        max-width="260"
        variant="outlined"
      />

      <v-spacer />
      <div class="text-center">
        <div class="text-h6 font-weight-bold">{{ totalValue }}</div>
        <div class="text-caption text-medium-emphasis">
          Total book value &middot; {{ totalCount }}
        </div>
      </div>
      <v-spacer />

      <v-btn
        color="primary"
        prepend-icon="mdi-plus"
        text="Add Location"
        @click="openAdd"
      />
    </div>

    <div class="chart-area">
      <AssetTreemap :location="selected" />
    </div>
  </div>

  <v-dialog v-model="dialog" max-width="420">
    <v-card>
      <v-card-title>Add Location</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="newName"
          autofocus
          density="compact"
          label="Location name"
          @keyup.enter="addLocation"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" @click="dialog = false" />
        <v-btn color="primary" text="Add" @click="addLocation" />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.assets-treemap-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}

.filters {
  flex: 0 0 auto;
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.chart-area {
  flex: 1 1 auto;
  min-height: 0;
}
</style>
