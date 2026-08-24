<template>
  <div class="barcodes-page">
    <div class="bc-header">
      <h2 class="bc-title">Barcodes</h2>

      <div class="bc-header-right">
        <v-text-field
          v-model="query"
          density="compact"
          hide-details
          placeholder="Search barcodes..."
          prepend-inner-icon="mdi-magnify"
          style="max-width: 260px"
          variant="solo"
        />

        <v-btn
          color="primary"
          size="small"
          @click="isNewOpen = true"
        >
          <v-icon start>mdi-plus</v-icon>
          New
        </v-btn>
      </div>
    </div>

    <div ref="scrollBody" class="bc-table-wrap">
      <v-table class="bc-table" density="compact">
        <thead>
          <tr>
            <th style="width: 36px">
              <input v-model="allChecked" type="checkbox">
            </th>

            <th style="width: 52px">Qty</th>
            <th>Product</th>
            <th style="min-width: 120px">Barcode</th>
            <th style="width: 80px">Preview</th>
            <th>Created</th>
            <th>SKU</th>
            <th>Status</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="r in visibleItems" :key="r.id" :class="{ 'bc-sel': checked.has(r.id) }">
            <td class="bc-check-cell">
              <input :checked="checked.has(r.id)" type="checkbox" @change="toggleCheck(r.id)">
            </td>

            <td class="bc-qty-cell">
              <input
                v-model.number="printQty[r.id]"
                class="bc-qty"
                :disabled="!checked.has(r.id)"
                max="100"
                min="1"
                type="number"
              >
            </td>

            <td>{{ r.product }}</td>
            <td class="bc-mono">{{ r.code }}</td>

            <td>
              <canvas :ref="(el) => setCanvas(el, r.code)" class="bc-canvas" />
            </td>

            <td>{{ formatDate(r.createdAt) }}</td>
            <td>{{ r.sku }}</td>

            <td>
              <v-chip :color="r.status === 'Active' ? 'success' : 'error'" size="x-small" :text="r.status" />
            </td>
          </tr>
        </tbody>
      </v-table>

      <div v-if="loading && hasMore" class="bc-load-more">
        <v-progress-circular color="primary" indeterminate size="20" />
      </div>

      <div ref="sentinel" class="bc-sentinel" />
    </div>

    <v-dialog v-model="isNewOpen" max-width="480" scroll="keep">
      <v-card>
        <v-card-title>New Barcode</v-card-title>

        <v-card-text>
          <v-text-field
            v-model="form.product"
            class="mb-2"
            density="compact"
            label="Product"
            prepend-inner-icon="mdi-package-variant"
            variant="outlined"
          />

          <v-text-field
            v-model="form.sku"
            class="mb-2"
            density="compact"
            label="SKU"
            prepend-inner-icon="mdi-identifier"
            variant="outlined"
          />

          <v-select
            v-model="form.status"
            density="compact"
            :items="['Active', 'Inactive']"
            label="Status"
            prepend-inner-icon="mdi-information"
            variant="outlined"
          />
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="isNewOpen = false">Cancel</v-btn>
          <v-btn color="primary" @click="addItem">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'

  const statuses = ['Active', 'Inactive']
  const products = ['Cotton T-Shirt', 'Denim Jeans', 'Silk Scarf', 'Wool Sweater', 'Linen Pants', 'Jersey Hoodie', 'Chinos', 'Blazer', 'Polo Shirt', 'Joggers']
  const skuPrefixes = ['TS', 'JN', 'SC', 'SW', 'LP', 'HD', 'CH', 'BZ', 'PS', 'JG']

  const pick = (a: string[]) => a[Math.floor(Math.random() * a.length)]
  const pageSize = 50

  const isNewOpen = ref(false)
  const query = ref('')
  const loading = ref(false)
  const loaded = ref(pageSize)
  const checked = ref(new Set<number>())
  const printQty = reactive<Record<number, number>>({})
  const form = ref({ product: '', sku: '', status: 'Active' })
  const scrollBody = ref<HTMLElement | null>(null)
  const sentinel = ref<HTMLElement | null>(null)
  let observer: IntersectionObserver | null = null

  function formatDate (d: Date) {
    return new Date(d).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
  }

  function ean13 () {
    const digits = Array.from({ length: 12 }, () => Math.floor(Math.random() * 10))
    const check = (10 - (digits.reduce((s, d, i) => s + d * (i % 2 ? 3 : 1), 0) % 10)) % 10
    return digits.join('') + check
  }

  const allItems = Array.from({ length: 500 }, (_, i) => {
    const item = {
      id: 7001 + i,
      product: pick(products),
      code: ean13(),
      createdAt: new Date(Date.now() - Math.random() * 180 * 86_400_000),
      sku: `${pick(skuPrefixes)}-${String(i + 1).padStart(4, '0')}`,
      status: pick(statuses),
    }
    printQty[item.id] = 1
    return item
  })

  const filteredItems = computed(() => {
    if (!query.value) return allItems
    const q = query.value.toLowerCase()
    return allItems.filter(r => [r.product, r.code, r.sku, r.status, String(r.id)].some(v => v.toLowerCase().includes(q)))
  })

  const visibleItems = computed(() => filteredItems.value.slice(0, loaded.value))
  const hasMore = computed(() => loaded.value < filteredItems.value.length)
  const allChecked = computed({
    get: () => visibleItems.value.length > 0 && visibleItems.value.every(r => checked.value.has(r.id)),
    set: v => v
      ? (() => {
        for (const r of visibleItems.value) checked.value.add(r.id)
      })()
      : checked.value.clear(),
  })

  function toggleCheck (id: number) {
    checked.value.has(id) ? checked.value.delete(id) : checked.value.add(id)
  }

  function setCanvas (el: HTMLCanvasElement | null, code: string) {
    if (!el) return
    drawBarcode(el, code, 120, 30)
  }

  function drawBarcode (canvas: HTMLCanvasElement, code: string, w: number, h: number) {
    canvas.width = w
    canvas.height = h
    const ctx = canvas.getContext('2d')!
    ctx.fillStyle = '#fff'
    ctx.fillRect(0, 0, w, h)
    ctx.fillStyle = '#000'
    const barW = w / 95
    const pattern = encodeEAN13(code)
    let x = 0
    for (const bar of pattern) {
      if (bar) ctx.fillRect(x, 0, barW, h)
      x += barW
    }
  }

  function encodeEAN13 (code: string) {
    const bars: number[] = []
    const digits = code.split('').map(Number)
    const L = ['0001101', '0011001', '0010011', '0111101', '0100011', '0110001', '0101111', '0111011', '0110111', '0001011']
    const G = ['0100111', '0110011', '0011011', '0100001', '0011101', '0111001', '0000101', '0010001', '0001001', '0010111']
    const R = ['1110010', '1100110', '1101100', '1000010', '1011100', '1001110', '1010000', '1000100', '1001000', '1110100']
    const enc = ['LLLLLL', 'LLGLGG', 'LLGGLG', 'LLGGGL', 'LGLLGG', 'LGGLLG', 'LGGGLL', 'LGLGLG', 'LGLGGL', 'LGGLGL']
    bars.push(1, 0, 1)
    const pat = enc[digits[0]]
    for (let i = 1; i <= 6; i++) {
      const p = pat[i - 1] === 'L' ? L[digits[i]] : G[digits[i]]
      for (const c of p) bars.push(Number(c))
    }
    bars.push(0, 1, 0, 1, 0)
    for (let i = 7; i <= 12; i++) {
      const p = R[digits[i]]
      for (const c of p) bars.push(Number(c))
    }
    bars.push(1, 0, 1)
    return bars
  }

  async function loadMore () {
    if (loading.value || !hasMore.value) return
    loading.value = true
    await new Promise(r => setTimeout(r, 300))
    loaded.value = Math.min(loaded.value + pageSize, filteredItems.value.length)
    loading.value = false
  }

  watch(query, () => {
    loaded.value = pageSize
  })

  onMounted(() => {
    observer = new IntersectionObserver(([e]) => {
      if (e.isIntersecting) loadMore()
    }, { root: scrollBody.value, rootMargin: '200px' })
    observer.observe(sentinel.value)
  })

  onUnmounted(() => observer?.disconnect())

  function addItem () {
    if (!form.value.product || !form.value.sku) return
    const item = { id: allItems.length + 7001, product: form.value.product, code: ean13(), createdAt: new Date(), sku: form.value.sku, status: form.value.status }
    allItems.unshift(item)
    printQty[item.id] = 1
    form.value = { product: '', sku: '', status: 'Active' }
    isNewOpen.value = false
    loaded.value = Math.min(loaded.value + 1, filteredItems.value.length)
  }
</script>

<style scoped>
.barcodes-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 48px);
  overflow: hidden;
}

.bc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid #E8EAED;
  flex-shrink: 0;
}

.bc-title {
  font-size: 15px;
  font-weight: 500;
  margin: 0;
}

.bc-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bc-table-wrap {
  flex: 1;
  overflow-y: auto;
}

.bc-table {
  font-size: 13px;
}

.bc-table td {
  padding: 4px 8px !important;
}

.bc-check-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}

.bc-check-cell input {
  accent-color: #4285F4;
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.bc-qty-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}

.bc-qty {
  width: 36px;
  height: 1.6rem;
  text-align: center;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 0.8rem;
  padding: 0 2px;
  -moz-appearance: textfield;
  appearance: textfield;
}

.bc-qty::-webkit-inner-spin-button,
.bc-qty::-webkit-outer-spin-button {
  -webkit-appearance: none;
}

.bc-qty:disabled {
  opacity: 0.4;
  cursor: default;
}

.bc-mono {
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
}

.bc-canvas {
  display: block;
  height: 30px;
}

.bc-sel {
  background: rgba(66, 133, 244, 0.08) !important;
}

.bc-load-more {
  display: flex;
  justify-content: center;
  padding: 12px;
}

.bc-sentinel {
  height: 1px;
}
</style>
