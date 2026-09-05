<template>
  <div class="page-root">
    <v-card class="barcodes-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-barcode" />
        <span>Barcodes</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="table-search"
          clearable
          density="compact"
          hide-details
          label="Search product / code / barcode"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
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
      </v-card-title>
      <v-divider />

      <div class="barcodes-content">
        <div class="pa-4">
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

          <ServerSideTable
            ref="tableRef"
            :api-u-r-l="apiURL"
            class="barcodes-table"
            density="comfortable"
            :external-search="search ?? ''"
            flush
            :headers="headers"
            item-value="id"
            root-key="variants"
            selectable
            :show-search-icon="false"
            @loaded="onLoaded"
            @update:selected="selected = $event"
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
                variant="solo-filled"
              />
            </template>
          </ServerSideTable>
        </div>
      </div>
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

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBarcodes } from '@/composables/useBarcodes'
  import { API_BASE } from '@/config'
  import { useSettingsStore } from '@/stores/settings'
  import { formatMoney } from '@/utils/money'

  const { assignBarcodes, printLabels } = useBarcodes()
  const settings = useSettingsStore()

  const apiURL = `${API_BASE}/barcodes`
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

  const tableRef = ref(null)
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

  // Seed a default print quantity of 1 for every freshly loaded row.
  function onLoaded (items) {
    for (const v of items) {
      if (qtyById[v.id] == null) qtyById[v.id] = 1
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
      tableRef.value?.reload()
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
  })
</script>

<style scoped>
/* Fill the layout's flex-column scroll wrapper so the scroll lives inside the
   table, not the whole page. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.barcodes-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.barcodes-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.table-search {
  flex: 0 1 320px;
}
.bc-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.bc-qty :deep(input) {
  appearance: textfield;
  min-height: 30px;
  padding: 2px 6px;
  text-align: center;
}

.bc-qty :deep(.v-field) {
  --v-input-control-height: 30px;
}

.bc-qty :deep(input::-webkit-inner-spin-button),
.bc-qty :deep(input::-webkit-outer-spin-button) {
  margin: 0;
  appearance: none;
}
</style>
