<template>
  <div class="page-root">
    <div class="d-flex flex-wrap align-center ga-3 mb-4">
      <h2 class="text-h6">Barcodes</h2>
      <v-spacer />
      <v-text-field
        v-model="search"
        clearable
        density="compact"
        hide-details
        label="Search product / code / barcode"
        prepend-inner-icon="mdi-magnify"
        style="max-width: 320px"
        variant="outlined"
        @update:model-value="debouncedReload"
      />
      <v-btn
        :disabled="selected.length === 0 || assigning"
        :loading="assigning"
        prepend-icon="mdi-barcode"
        text="Generate barcodes"
        variant="tonal"
        @click="generate"
      />
      <v-btn
        color="primary"
        :disabled="selected.length === 0"
        prepend-icon="mdi-printer"
        text="Print PDF"
        variant="flat"
        @click="printDialog = true"
      />
    </div>

    <v-alert
      v-if="message"
      class="mb-3"
      closable
      density="compact"
      :type="messageType"
      variant="tonal"
      @click:close="message = ''"
    >
      {{ message }}
    </v-alert>

    <v-card flat>
      <v-data-table-server
        v-model="selected"
        density="compact"
        :headers="headers"
        item-value="id"
        :items="rows"
        :items-length="total"
        :items-per-page="pageSize"
        :items-per-page-options="[14, 25, 50, 100]"
        :loading="loading"
        :page="page + 1"
        show-select
        @update:options="onOptions"
      >
        <template #item.product="{ item }">
          <div class="text-body-2">{{ item.product_name }}</div>
          <div class="text-caption text-medium-emphasis">{{ item.product_code }}</div>
        </template>
        <template #item.variant="{ item }">
          <span v-if="item.color_name || item.size_name">
            {{ [item.color_name, item.size_name].filter(Boolean).join(' / ') }}
          </span>
          <span v-else class="text-medium-emphasis">—</span>
        </template>
        <template #item.barcode="{ item }">
          <span class="bc-mono">{{ item.barcode }}</span>
          <v-chip
            v-if="!isEan13(item.barcode)"
            class="ms-2"
            color="warning"
            size="x-small"
            text="no barcode"
          />
        </template>
        <template #item.price="{ item }">
          {{ item.price ? formatMoney(item.price) : '—' }}
        </template>
        <template #item.qty="{ item }">
          <v-text-field
            v-model.number="qtyById[item.id]"
            class="bc-qty"
            density="compact"
            hide-details
            min="1"
            style="width: 72px"
            type="number"
            variant="outlined"
          />
        </template>
      </v-data-table-server>
    </v-card>

    <!-- Print options -->
    <v-dialog v-model="printDialog" max-width="480">
      <v-card>
        <v-card-title class="text-h6">Print {{ selected.length }} item{{ selected.length === 1 ? '' : 's' }}</v-card-title>
        <v-card-text>
          <v-checkbox
            v-model="separatorEnabled"
            density="compact"
            hide-details
            label="Separator page before each item (adds side borders to labels)"
          />
          <v-expand-transition>
            <div v-show="separatorEnabled" class="ms-8 mt-1">
              <div class="text-caption text-medium-emphasis mb-1">Fields to show on the separator page</div>
              <v-checkbox
                v-for="f in separatorFieldOptions"
                :key="f.value"
                v-model="separatorFields"
                density="compact"
                hide-details
                :label="f.title"
                :value="f.value"
              />
            </div>
          </v-expand-transition>

          <v-divider class="my-3" />

          <div class="text-caption text-medium-emphasis mb-1">Label size (points, 72 = 1 inch)</div>
          <div class="d-flex ga-3">
            <v-text-field
              v-model.number="labelWidth"
              density="compact"
              hide-details
              label="Width"
              type="number"
              variant="outlined"
            />
            <v-text-field
              v-model.number="labelHeight"
              density="compact"
              hide-details
              label="Height"
              type="number"
              variant="outlined"
            />
          </div>
          <div class="text-caption text-medium-emphasis mt-1">
            Default comes from Settings ({{ settings.barcodeLabelWidth }} × {{ settings.barcodeLabelHeight }}).
          </div>

          <v-alert
            v-if="printError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >{{ printError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="printDialog = false" />
          <v-btn
            color="primary"
            :loading="printing"
            text="Generate PDF"
            variant="flat"
            @click="doPrint"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
  import { onMounted, reactive, ref } from 'vue'
  import { useBarcodes } from '@/composables/useBarcodes'
  import { useSettingsStore } from '@/stores/settings'
  import { formatMoney } from '@/utils/money'

  const { listVariants, assignBarcodes, printLabels } = useBarcodes()
  const settings = useSettingsStore()

  const headers = [
    { title: 'Product', key: 'product', sortable: false },
    { title: 'Variant', key: 'variant', sortable: false },
    { title: 'Barcode', key: 'barcode', sortable: false },
    { title: 'Price', key: 'price', align: 'end', sortable: false },
    { title: 'Qty to print', key: 'qty', align: 'center', sortable: false, width: 110 },
  ]

  const separatorFieldOptions = [
    { title: 'Name', value: 'name' },
    { title: 'Quantity', value: 'qty' },
    { title: 'Brand', value: 'brand' },
    { title: 'Color', value: 'color' },
    { title: 'Size', value: 'size' },
  ]

  const rows = ref([])
  const total = ref(0)
  const loading = ref(false)
  const page = ref(0)
  const pageSize = ref(14)
  const search = ref('')

  const selected = ref([])
  const qtyById = reactive({})

  const message = ref('')
  const messageType = ref('success')
  const assigning = ref(false)

  const printDialog = ref(false)
  const printing = ref(false)
  const printError = ref('')
  const separatorEnabled = ref(false)
  const separatorFields = ref(['name', 'qty'])
  const labelWidth = ref(0)
  const labelHeight = ref(0)

  function isEan13 (v) {
    return typeof v === 'string' && /^\d{13}$/.test(v)
  }

  let searchTimer
  function debouncedReload () {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      page.value = 0
      reload()
    }, 300)
  }

  async function reload () {
    loading.value = true
    try {
      const { variants, total: t } = await listVariants({
        page: page.value,
        pageSize: pageSize.value,
        search: search.value || '',
      })
      rows.value = variants
      total.value = t
      for (const v of variants) {
        if (qtyById[v.id] == null) {
          qtyById[v.id] = 1
        }
      }
    } catch (error) {
      message.value = error.message
      messageType.value = 'error'
    } finally {
      loading.value = false
    }
  }

  function onOptions ({ page: p, itemsPerPage }) {
    const nextPage = (p ?? 1) - 1
    const changed = nextPage !== page.value || itemsPerPage !== pageSize.value
    page.value = nextPage
    pageSize.value = itemsPerPage
    if (changed) {
      reload()
    }
  }

  async function generate () {
    assigning.value = true
    message.value = ''
    try {
      const res = await assignBarcodes(selected.value)
      const n = res?.updated?.length ?? 0
      message.value = n
        ? `Generated ${n} barcode${n === 1 ? '' : 's'}.`
        : 'All selected items already had a barcode.'
      messageType.value = 'success'
      await reload()
    } catch (error) {
      message.value = error.message
      messageType.value = 'error'
    } finally {
      assigning.value = false
    }
  }

  async function doPrint () {
    printing.value = true
    printError.value = ''
    try {
      const items = selected.value.map(id => ({ variant_id: id, qty: Math.max(1, Number(qtyById[id]) || 1) }))
      const blob = await printLabels({
        items,
        separator: { enabled: separatorEnabled.value, fields: separatorFields.value },
        label: { width: Number(labelWidth.value) || 0, height: Number(labelHeight.value) || 0 },
      })
      const url = URL.createObjectURL(blob)
      window.open(url, '_blank')
      setTimeout(() => URL.revokeObjectURL(url), 60_000)
      printDialog.value = false
    } catch (error) {
      printError.value = error.message
    } finally {
      printing.value = false
    }
  }

  onMounted(async () => {
    await settings.init()
    labelWidth.value = settings.barcodeLabelWidth
    labelHeight.value = settings.barcodeLabelHeight
    reload()
  })
</script>

<style scoped>
.page-root {
  flex: 0 0 auto;
}
.bc-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.bc-qty :deep(input) {
  text-align: center;
}
</style>
