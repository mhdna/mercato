<template>
  <v-dialog v-model="dialog" max-width="560" :retain-focus="false" scrollable>
    <v-card>
      <v-card-title class="d-flex align-center">
        <span class="text-h6">New Product</span>
        <v-spacer />
        <v-btn icon="mdi-close" size="small" variant="text" @click="dialog = false" />
      </v-card-title>
      <v-divider />
      <v-card-text class="py-4">
        <NewProductForm @created="onProductCreated" />
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="editDialog" max-width="760" :retain-focus="false" scrollable>
    <EditProductForm
      v-if="editDialog"
      :product-id="editId"
      @close="editDialog = false"
      @deleted="onProductChanged"
      @saved="onProductChanged"
    />
  </v-dialog>

  <div class="page-root">
    <v-card class="products-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-package-variant" />
        <span>Products</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="products-search"
          clearable
          density="compact"
          hide-details
          label="Search products"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-badge
          color="primary"
          :content="activeFilterCount"
          :model-value="activeFilterCount > 0"
        >
          <v-btn
            prepend-icon="mdi-filter-variant"
            text="Filters"
            variant="tonal"
            @click="filterMenu = true"
          />
        </v-badge>
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Product"
          variant="flat"
          @click="dialog = true"
        />
      </v-card-title>
      <v-divider />

      <v-dialog v-model="filterMenu" max-width="640" scrollable>
        <v-card>
          <v-card-title class="d-flex align-center ga-2">
            <v-icon icon="mdi-filter-variant" size="20" />
            <span class="text-h6">Filter Products</span>
            <v-chip
              v-if="activeFilterCount > 0"
              class="ms-1"
              color="primary"
              size="small"
              :text="`${activeFilterCount} active`"
              variant="tonal"
            />
            <v-spacer />
            <v-btn icon="mdi-close" size="small" variant="text" @click="filterMenu = false" />
          </v-card-title>
          <v-divider />

          <v-card-text class="py-4">
            <div class="text-caption text-medium-emphasis mb-1">Status</div>
            <v-btn-toggle
              v-model="statusFilter"
              class="mb-5"
              color="primary"
              density="compact"
              divided
              mandatory
              variant="outlined"
            >
              <v-btn value="all">All</v-btn>
              <v-btn value="active">Active</v-btn>
              <v-btn value="inactive">Inactive</v-btn>
            </v-btn-toggle>

            <v-switch
              v-model="hasVariants"
              class="mb-2"
              color="primary"
              density="compact"
              hide-details
              label="Has variants"
            />

            <div class="text-caption text-medium-emphasis mb-1">Created between</div>
            <div class="d-flex flex-wrap ga-3 mb-5">
              <v-text-field
                v-model="createdFrom"
                class="date-field"
                clearable
                density="compact"
                hide-details
                label="From"
                type="date"
                variant="outlined"
              />
              <v-text-field
                v-model="createdTo"
                class="date-field"
                clearable
                density="compact"
                hide-details
                label="To"
                type="date"
                variant="outlined"
              />
            </div>

            <template v-if="attributeTypes.length > 0">
              <div class="text-caption text-medium-emphasis mb-1">Attributes</div>
              <div class="d-flex flex-wrap ga-3">
                <v-select
                  v-for="t in attributeTypes"
                  :key="t.name"
                  v-model="attrSelections[t.name]"
                  chips
                  class="attr-select"
                  clearable
                  closable-chips
                  density="compact"
                  hide-details
                  item-title="value"
                  item-value="id"
                  :items="valuesByType[t.name] || []"
                  :label="labelize(t.name)"
                  multiple
                  variant="outlined"
                />
              </div>
            </template>
          </v-card-text>

          <v-divider />
          <v-card-actions class="px-4">
            <v-btn
              :disabled="activeFilterCount === 0"
              text="Clear all"
              variant="text"
              @click="clearFilters"
            />
            <v-spacer />
            <v-btn color="primary" text="Done" variant="flat" @click="filterMenu = false" />
          </v-card-actions>
        </v-card>
      </v-dialog>

      <ProductsTable
        ref="productsTableRef"
        :filters="filters"
        :search="search"
        @edit="openEdit"
      />
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import EditProductForm from '@/components/Forms/EditProductForm.vue'
  import NewProductForm from '@/components/Forms/NewProductForm.vue'
  import ProductsTable from '@/components/ProductsTable.vue'
  import { useAttributes } from '@/composables/useAttributes'

  const dialog = ref(false)
  const search = ref('')
  const productsTableRef = ref(null)

  const editDialog = ref(false)
  const editId = ref(null)

  function openEdit (product) {
    editId.value = product.id
    editDialog.value = true
  }

  function onProductCreated () {
    dialog.value = false
    productsTableRef.value?.reload()
  }

  function onProductChanged () {
    editDialog.value = false
    productsTableRef.value?.reload()
  }

  // ---- filters -----------------------------------------------------------
  const { types, valuesByType, fetchAll } = useAttributes()
  onMounted(() => fetchAll())

  const attributeTypes = computed(() => types.value)
  const attrSelections = reactive({})

  const filterMenu = ref(false)
  const statusFilter = ref('all')
  const hasVariants = ref(false)
  const createdFrom = ref(null)
  const createdTo = ref(null)

  function labelize (name) {
    if (!name) return ''
    return name.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
  }

  const selectedAttributeIds = computed(() =>
    Object.values(attrSelections).flat().filter(Boolean),
  )

  const filters = computed(() => ({
    is_active: statusFilter.value === 'all' ? null : statusFilter.value === 'active',
    has_variants: hasVariants.value ? true : null,
    created_from: createdFrom.value || null,
    created_to: createdTo.value || null,
    attribute_value_ids: selectedAttributeIds.value,
  }))

  const activeFilterCount = computed(() => {
    let n = 0
    if (statusFilter.value !== 'all') n++
    if (hasVariants.value) n++
    if (createdFrom.value) n++
    if (createdTo.value) n++
    n += selectedAttributeIds.value.length
    return n
  })

  function clearFilters () {
    statusFilter.value = 'all'
    hasVariants.value = false
    createdFrom.value = null
    createdTo.value = null
    for (const key of Object.keys(attrSelections)) attrSelections[key] = []
  }
</script>

<style scoped>
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.products-card {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
}
.products-search {
  max-width: 320px;
  min-width: 220px;
}
.date-field {
  max-width: 170px;
}
.attr-select {
  min-width: 200px;
  max-width: 240px;
}
</style>
